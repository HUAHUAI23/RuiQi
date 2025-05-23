package internal

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// 原始的bufio.Scanner实现
func getHeaderValueOld(headers []byte, targetHeader string) (string, error) {
	s := bufio.NewScanner(bytes.NewReader(headers))
	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue
		}

		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			continue
		}

		key, value := bytes.TrimSpace(kv[0]), bytes.TrimSpace(kv[1])
		if strings.EqualFold(string(key), targetHeader) {
			return string(value), nil
		}
	}
	return "", nil
}

func TestGetHeaderValue(t *testing.T) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100
Content-Type: application/json
Content-Length: 1234
Authorization: Bearer token123`)

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"host", "host", "example.com"},
		{"Host", "Host", "example.com"},
		{"HOST", "HOST", "example.com"},
		{"x-real-ip", "x-real-ip", "192.168.1.100"},
		{"X-Real-IP", "X-Real-IP", "192.168.1.100"},
		{"x-forwarded-for", "x-forwarded-for", "10.0.0.1, 192.168.1.100"},
		{"content-type", "content-type", "application/json"},
		{"not-exist", "not-exist", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试新实现
			result, err := getHeaderValue(testHeaders, tt.header)
			if err != nil {
				t.Errorf("getHeaderValue() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("getHeaderValue() = %v, want %v", result, tt.expected)
			}

			// 测试旧实现对比
			oldResult, err := getHeaderValueOld(testHeaders, tt.header)
			if err != nil {
				t.Errorf("getHeaderValueOld() error = %v", err)
				return
			}
			if oldResult != result {
				t.Errorf("结果不一致: new=%v, old=%v", result, oldResult)
			}
		})
	}
}

func BenchmarkGetHeaderValue(b *testing.B) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100, 172.16.0.1
X-Forwarded-Proto: https
X-Cluster-Client-IP: 10.0.0.1
Content-Type: application/json; charset=utf-8
Content-Length: 1234
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
Cache-Control: no-cache
Accept: application/json, text/plain, */*
Accept-Language: en-US,en;q=0.9,zh-CN;q=0.8
Accept-Encoding: gzip, deflate, br`)

	b.Run("New_Implementation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValue(testHeaders, "x-forwarded-for")
		}
	})

	b.Run("Old_Implementation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "x-forwarded-for")
		}
	})

	// 模拟真实场景：查找多个header（如getRealClientIP函数中的使用）
	headers := []string{
		"x-forwarded-for", "x-real-ip", "true-client-ip", "cf-connecting-ip",
		"fastly-client-ip", "x-client-ip", "x-original-forwarded-for",
		"forwarded", "x-cluster-client-ip",
	}

	b.Run("New_Multiple_Headers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, header := range headers {
				_, _ = getHeaderValue(testHeaders, header)
			}
		}
	})

	b.Run("Old_Multiple_Headers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, header := range headers {
				_, _ = getHeaderValueOld(testHeaders, header)
			}
		}
	})
}

func BenchmarkGetHeaderValueWorstCase(b *testing.B) {
	// 创建一个很长的header列表，目标header在最后
	var headerBuilder strings.Builder
	for i := 0; i < 20; i++ {
		headerBuilder.WriteString("Header-")
		headerBuilder.WriteString(strings.Repeat("X", 10))
		headerBuilder.WriteString(": value")
		headerBuilder.WriteString(strings.Repeat("Y", 50))
		headerBuilder.WriteByte('\n')
	}
	headerBuilder.WriteString("Target-Header: found-value\n")

	testHeaders := []byte(headerBuilder.String())

	b.Run("New_Worst_Case", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValue(testHeaders, "target-header")
		}
	})

	b.Run("Old_Worst_Case", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "target-header")
		}
	})
}

// 原始实现（用于性能对比）
func getRealClientIPOriginal(req *applicationRequest) string {
	if req == nil {
		return ""
	}

	// 按优先级尝试不同的头部
	headers := []string{
		"x-forwarded-for",  // 最常用，链式格式
		"x-real-ip",        // Nginx常用
		"true-client-ip",   // Akamai
		"cf-connecting-ip", // Cloudflare
		"fastly-client-ip", // Fastly
		"x-client-ip",      // 通用
		"x-original-forwarded-for",
		"forwarded", // 标准头部
		"x-cluster-client-ip",
	}

	// 尝试从各个头部获取IP
	for _, header := range headers {
		if value, err := getHeaderValue(req.Headers, header); err == nil && value != "" {
			// 对于X-Forwarded-For和类似的链式格式，提取第一个IP
			if header == "x-forwarded-for" || header == "x-original-forwarded-for" {
				ips := strings.Split(value, ",")
				if len(ips) > 0 {
					ip := strings.TrimSpace(ips[0])
					if ip != "" {
						return ip
					}
				}
			} else if header == "forwarded" { // 对于Forwarded头部，需要特殊处理
				// 解析Forwarded头部，格式如：for=client;proto=https;by=proxy
				parts := strings.Split(value, ";")
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "for" {
						// 去除可能的引号和IPv6方括号
						ip := strings.TrimSpace(kv[1])
						ip = strings.Trim(ip, "\"")

						// 处理IPv6地址特殊格式
						if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
							ip = ip[1 : len(ip)-1]
						}

						if ip != "" {
							return ip
						}
					}
				}
			} else { // 其他头部直接返回值
				ip := strings.TrimSpace(value)
				if ip != "" {
					return ip
				}
			}
		}
	}

	// 如果所有头部都没有，返回源IP
	if req.SrcIp.IsValid() {
		return req.SrcIp.String()
	}

	return ""
}

func BenchmarkGetRealClientIPOptimization(b *testing.B) {
	// 测试不同场景的header
	testCases := []struct {
		name    string
		headers []byte
	}{
		{
			name: "XForwardedFor_First",
			headers: []byte(`X-Forwarded-For: 192.168.1.100, 10.0.0.1
Host: example.com
User-Agent: Mozilla/5.0`),
		},
		{
			name: "XForwardedFor_Middle",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100, 10.0.0.1
User-Agent: Mozilla/5.0`),
		},
		{
			name: "XRealIP_Only",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.100
User-Agent: Mozilla/5.0`),
		},
		{
			name: "CloudFlare_Only",
			headers: []byte(`Host: example.com
CF-Connecting-IP: 192.168.1.100
User-Agent: Mozilla/5.0`),
		},
		{
			name: "No_Client_IP_Headers",
			headers: []byte(`Host: example.com
User-Agent: Mozilla/5.0
Content-Type: application/json`),
		},
		{
			name: "Many_Headers_XFF_Last",
			headers: []byte(`Host: example.com
User-Agent: Mozilla/5.0
Accept: application/json
Content-Type: application/json
Cache-Control: no-cache
Authorization: Bearer token
X-Forwarded-For: 192.168.1.100, 10.0.0.1`),
		},
	}

	for _, tc := range testCases {
		req := &applicationRequest{Headers: tc.headers}

		b.Run(tc.name+"_Original", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIPOriginal(req)
			}
		})

		b.Run(tc.name+"_Optimized", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIP(req)
			}
		})
	}
}
