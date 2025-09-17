package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	_ "github.com/gookit/goutil/dump"
)

// 允许的文件大小，默认999GB，相当于无限制
const SIZE_LIMIT = 1024 * 1024 * 1024 * 999

var blobToRaw = regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$`)

// 全局变量：被阻止的内容类型
var blockedContentTypes = map[string]bool{
	"text/html":             true,
	"application/xhtml+xml": true,
	"text/xml":              true,
	"application/xml":       true,
}

// GitHubProxyHandler GitHub代理处理器
func GitHubProxyHandler(c *fiber.Ctx) error {
	rawPath := c.Params("*")
	cfg := c.Locals("config").(Config)

	for strings.HasPrefix(rawPath, "/") {
		rawPath = strings.TrimPrefix(rawPath, "/")
	}

	// 自动补全协议头
	if !strings.HasPrefix(rawPath, "https://") {
		if strings.HasPrefix(rawPath, "http:/") || strings.HasPrefix(rawPath, "https:/") {
			rawPath = strings.Replace(rawPath, "http:/", "", 1)
			rawPath = strings.Replace(rawPath, "https:/", "", 1)
		}
		rawPath, _ = strings.CutPrefix(rawPath, "http://")
		rawPath = "https://" + rawPath
	}

	if !isValidUrl(&cfg, rawPath) {
		return c.Status(fiber.StatusForbidden).SendString("无效请求")
	}

	// 将blob链接转换为raw链接
	if blobToRaw.MatchString(rawPath) {
		rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
	}

	return ProxyGitHubRequest(c, rawPath)
}

// ProxyGitHubRequest 代理GitHub请求
func ProxyGitHubRequest(c *fiber.Ctx, u string) error {
	return proxyGitHubWithRedirect(c, u, 0)
}

// proxyGitHubWithRedirect 带重定向的GitHub代理请求
func proxyGitHubWithRedirect(c *fiber.Ctx, u string, redirectCount int) error {
	const maxRedirects = 20
	if redirectCount > maxRedirects {
		return c.Status(fiber.StatusLoopDetected).SendString("重定向次数过多，可能存在循环重定向")
	}

	cfg := c.Locals("config").(Config)

	agent := getFiberAgent(u)
	agent.Request().Header.SetMethod(c.Method())
	agent.Body(c.Request().Body())

	// 重定向之后的Location里面链接可能带有查询参数，Go的框架一般用QueryString表示URL里面问号之后的键值对
	// 这些查询参数需要传给客户端用于发到目标链接，内容可能会涉及Token之类的
	if queryStr := c.Request().URI().QueryString(); len(queryStr) > 0 {
		agent.QueryStringBytes(queryStr)
	}

	// 复制请求头
	for key, values := range c.GetReqHeaders() {
		for _, value := range values {
			agent.Add(key, value)
		}
	}

	agent.Request().Header.Del("Host")
	resp := fiber.AcquireResponse()
	defer func() {
		if err := resp.CloseBodyStream(); err != nil {
			log.Errorf("关闭代理响应体失败: %v\n", err)
		}
		resp.ConnectionClose()
		if err := c.Response().CloseBodyStream(); err != nil {
			log.Errorf("关闭响应体失败: %v\n", err)
		}
		fiber.ReleaseResponse(resp)
		fiber.ReleaseAgent(agent)
	}()

	agent.ReadBufferSize = 1024 * 64
	agent.WriteBufferSize = 1024 * 64
	agent.StreamResponseBody = true

	// 在国内用代理做加速
	if cfg.ProxyEnabled {
		if dialFn := adaptDialer(cfg.Proxy); dialFn != nil {
			agent.Dial = dialFn
		}
	}

	if err := agent.Do(agent.Request(), resp); err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Server Error: %v\n", err))
	}

	// 检查文件大小限制
	if size := resp.Header.ContentLength(); size > SIZE_LIMIT {
		return c.Status(fiber.StatusRequestEntityTooLarge).SendString(fmt.Sprintf("文件过大，限制大小: %d MB", SIZE_LIMIT/(1024*1024)))
	}

	// 清理安全相关的头
	resp.Header.Del("Content-Security-Policy")
	resp.Header.Del("Referrer-Policy")
	resp.Header.Del("Strict-Transport-Security")

	// 复制其他响应头
	resp.Header.CopyTo(&c.Response().Header)

	// 处理重定向
	// 重定向之后的内容会自动发送 text/html 的内容类型头信息
	// 这样会被检测规则给拒绝，所以要先把text/html
	if location := string(resp.Header.Peek("Location")); location != "" {
		if isValidUrl(&cfg, u) {
			c.Set(fiber.HeaderLocation, "/"+location)
			return c.SendStatus(resp.StatusCode())
		} else {
			// 递归重定向，最大不超过20次重定向
			if err := proxyGitHubWithRedirect(c, location, redirectCount+1); err != nil {
				return err
			}

			return nil
		}
	}

	// 检查并处理被阻止的内容类型
	if c.Method() == "GET" {
		if contentType := string(resp.Header.ContentType()); blockedContentTypes[strings.ToLower(strings.Split(contentType, ";")[0])] {
			return c.Status(fiber.StatusForbidden).JSON(map[string]string{
				"error":   "Content type not allowed",
				"message": "检测到网页类型，本服务不支持加速网页，请检查您的链接是否正确。",
			})
		}
	}

	// 获取真实域名
	realHost := c.Hostname()
	if !strings.HasPrefix(realHost, "http://") && !strings.HasPrefix(realHost, "https://") {
		realHost = "https://" + realHost
	}

	var processedBody io.Reader = resp.BodyStream()
	var processedSize int64 = 0
	var err error

	// 智能处理.sh .ps1 .py文件
	if strings.HasSuffix(strings.ToLower(u), ".sh") || strings.HasSuffix(strings.ToLower(u), ".ps1") || strings.HasSuffix(strings.ToLower(u), ".py") {
		isGzipCompressed := string(resp.Header.ContentEncoding()) == "gzip"

		processedBody, processedSize, err = ProcessSmart(resp.BodyStream(), isGzipCompressed, realHost)
		if err != nil {
			fmt.Printf("智能处理失败，回退到直接代理: %v\n", err)
			processedBody = resp.BodyStream()
			processedSize = 0
		}

		// 智能设置响应头
		if processedSize > 0 {
			resp.Header.Del("Content-Length")
			resp.Header.Del("Content-Encoding")
			resp.Header.Set("Transfer-Encoding", "chunked")
		}
	}

	// 输出处理后的内容
	_, err = io.Copy(c.Response().BodyWriter(), processedBody)

	return err
}
