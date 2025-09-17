package main

// 代码来源：
// https://github.com/sky22333/hubproxy/blob/main/src/utils/proxy_shell.go

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// GitHub URL正则表达式
var githubRegex = regexp.MustCompile(`https?://(?:github\.com|raw\.githubusercontent\.com|raw\.github\.com|gist\.githubusercontent\.com|gist\.github\.com|api\.github\.com)[^\s'"]+`)

// ProcessSmart Shell脚本智能处理函数
func ProcessSmart(input io.Reader, isCompressed bool, host string) (io.Reader, int64, error) {
	content, err := readShellContent(input, isCompressed)
	if err != nil {
		return nil, 0, fmt.Errorf("内容读取失败: %v", err)
	}

	if len(content) == 0 {
		return strings.NewReader(""), 0, nil
	}

	if len(content) > 10*1024*1024 {
		return strings.NewReader(content), int64(len(content)), nil
	}

	if !strings.Contains(content, "github.com") && !strings.Contains(content, "githubusercontent.com") {
		return strings.NewReader(content), int64(len(content)), nil
	}

	processed := processGitHubURLs(content, host)

	return strings.NewReader(processed), int64(len(processed)), nil
}

func readShellContent(input io.Reader, isCompressed bool) (string, error) {
	var reader io.Reader = input

	if isCompressed {
		peek := make([]byte, 2)
		n, err := input.Read(peek)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("读取数据失败: %v", err)
		}

		if n >= 2 && peek[0] == 0x1f && peek[1] == 0x8b {
			combinedReader := io.MultiReader(bytes.NewReader(peek[:n]), input)
			gzReader, err := gzip.NewReader(combinedReader)
			if err != nil {
				return "", fmt.Errorf("gzip解压失败: %v", err)
			}
			defer gzReader.Close()
			reader = gzReader
		} else {
			reader = io.MultiReader(bytes.NewReader(peek[:n]), input)
		}
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取内容失败: %v", err)
	}

	return string(data), nil
}

func processGitHubURLs(content, host string) string {
	return githubRegex.ReplaceAllStringFunc(content, func(url string) string {
		return transformURL(url, host)
	})
}
