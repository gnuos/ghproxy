package main

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

func getFiberAgent(uri string) *fiber.Agent {
	a := fiber.Get(uri).Reuse().InsecureSkipVerify().UserAgent(UA)
	a.Request().Header.Set(fiber.HeaderAccept, "*")
	a.Request().Header.Set(fiber.HeaderConnection, "Keep-Alive")

	return a
}

// 暂时预留一个基于req库的客户端
// req库的客户端对象也是可以复用的
func getReqClient() *req.Client {
	return req.C().DevMode().EnableInsecureSkipVerify().SetUserAgent(UA)
}

func adaptDialer(proxy string) fasthttp.DialFunc {
	var dialFn fasthttp.DialFunc
	proxyType := "no"
	proxyAddr := proxy

	if proxyAddr == "" {
		proxyType = "system"
	} else if strings.HasPrefix(proxyAddr, "http") {
		if u, ok := url.Parse(proxyAddr); ok == nil {
			proxyAddr = u.Host
			proxyType = "http"
		}
	} else if strings.HasPrefix(proxyAddr, "socks5://") {
		if _, ok := url.Parse(proxyAddr); ok == nil {
			proxyType = "socks5"
		}
	} else {
		if _, err := url.Parse("http://" + proxyAddr); err == nil {
			proxyType = "http"
		}
	}

	switch proxyType {
	case "system":
		{
			dialFn = fasthttpproxy.FasthttpProxyHTTPDialer()
			break
		}
	case "http":
		{
			dialFn = fasthttpproxy.FasthttpHTTPDialer(proxy)
			break
		}
	case "socks5":
		{
			dialFn = fasthttpproxy.FasthttpSocksDialer(proxy)
			break
		}
	default:
		{
			return nil
		}
	}

	return dialFn
}
