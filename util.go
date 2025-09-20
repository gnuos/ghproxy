package main

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

func isValidUrl(rules []string, uri string) bool {
	for _, rule := range rules {
		m, err := regexp.MatchString(rule, uri)
		if err != nil {
			log.Error(err)
			continue
		}

		if m {
			return m
		}
	}

	return false
}

// transformURL URL转换函数
func transformURL(url, host string) string {
	if strings.Contains(url, host) {
		return url
	}

	if strings.HasPrefix(url, "http://") {
		url = "https" + url[4:]
	} else if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "//") {
		url = "https://" + url
	}

	// 确保 host 有协议头
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	host = strings.TrimSuffix(host, "/")

	return host + "/" + url
}
