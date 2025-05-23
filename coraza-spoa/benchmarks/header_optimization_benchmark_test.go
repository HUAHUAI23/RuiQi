package benchmarks

import (
	"bufio"
	"bytes"
	"net/netip"
	"strings"
	"testing"
)

// 这些是从 internal 包导入的私有函数，需要通过接口访问
// 或者我们复制实现用于基准测试

// 原始的bufio.Scanner实现（用于性能对比）
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

// 优化后的实现（复制用于基准测试）
func getHeaderValueOptimized(headers []byte, targetHeader string) (string, error) {
	// 预处理目标头部为小写，避免重复转换
	targetLower := strings.ToLower(targetHeader)
	targetBytes := []byte(targetLower)
	targetLen := len(targetBytes)

	i := 0
	for i < len(headers) {
		// 找到行的开始
		lineStart := i

		// 找到行的结束（\n 或 \r\n）
		for i < len(headers) && headers[i] != '\n' && headers[i] != '\r' {
			i++
		}
		lineEnd := i

		// 跳过换行符
		if i < len(headers) && headers[i] == '\r' {
			i++
		}
		if i < len(headers) && headers[i] == '\n' {
			i++
		}

		// 跳过空行
		if lineEnd == lineStart {
			continue
		}

		// 在当前行中查找冒号
		colonPos := -1
		for j := lineStart; j < lineEnd; j++ {
			if headers[j] == ':' {
				colonPos = j
				break
			}
		}

		if colonPos == -1 {
			continue // 没有冒号，跳过这行
		}

		// 提取key（去除前后空格）
		keyStart := lineStart
		keyEnd := colonPos

		// 去除key前的空格
		for keyStart < keyEnd && (headers[keyStart] == ' ' || headers[keyStart] == '\t') {
			keyStart++
		}

		// 去除key后的空格
		for keyEnd > keyStart && (headers[keyEnd-1] == ' ' || headers[keyEnd-1] == '\t') {
			keyEnd--
		}

		keyLen := keyEnd - keyStart

		// 快速长度检查
		if keyLen != targetLen {
			continue
		}

		// 手动进行大小写不敏感比较
		match := true
		for j := 0; j < keyLen; j++ {
			c := headers[keyStart+j]
			// 转换为小写进行比较
			if c >= 'A' && c <= 'Z' {
				c = c + ('a' - 'A')
			}
			if c != targetBytes[j] {
				match = false
				break
			}
		}

		if !match {
			continue
		}

		// 找到匹配的header，提取value
		valueStart := colonPos + 1
		valueEnd := lineEnd

		// 去除value前的空格
		for valueStart < valueEnd && (headers[valueStart] == ' ' || headers[valueStart] == '\t') {
			valueStart++
		}

		// 去除value后的空格
		for valueEnd > valueStart && (headers[valueEnd-1] == ' ' || headers[valueEnd-1] == '\t') {
			valueEnd--
		}

		// 返回value
		if valueEnd > valueStart {
			return string(headers[valueStart:valueEnd]), nil
		}
		return "", nil
	}

	return "", nil
}

// ApplicationRequest 模拟结构体用于测试
type ApplicationRequest struct {
	Headers []byte
	SrcIp   netip.Addr
}

// 原始实现（用于性能对比）
func getRealClientIPOriginal(req *ApplicationRequest) string {
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
		if value, err := getHeaderValueOptimized(req.Headers, header); err == nil && value != "" {
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

// 优化后的getRealClientIP实现（模拟）
func getRealClientIPOptimized(req *ApplicationRequest) string {
	if req == nil {
		return ""
	}

	// 快速路径：直接查找最常用的X-Forwarded-For头部
	if value, err := getXForwardedForIP(req.Headers); err == nil && value != "" {
		return value
	}

	// 回退到原始逻辑
	return getRealClientIPOriginal(req)
}

// 专用的X-Forwarded-For解析函数
func getXForwardedForIP(headers []byte) (string, error) {
	target := []byte("x-forwarded-for")
	targetLen := len(target)

	i := 0
	for i < len(headers) {
		lineStart := i

		// 找到行结束
		for i < len(headers) && headers[i] != '\n' && headers[i] != '\r' {
			i++
		}
		lineEnd := i

		// 跳过换行符
		if i < len(headers) && headers[i] == '\r' {
			i++
		}
		if i < len(headers) && headers[i] == '\n' {
			i++
		}

		if lineEnd == lineStart {
			continue
		}

		// 查找冒号
		colonPos := -1
		for j := lineStart; j < lineEnd; j++ {
			if headers[j] == ':' {
				colonPos = j
				break
			}
		}

		if colonPos == -1 {
			continue
		}

		// 检查key
		keyStart := lineStart
		keyEnd := colonPos

		for keyStart < keyEnd && (headers[keyStart] == ' ' || headers[keyStart] == '\t') {
			keyStart++
		}
		for keyEnd > keyStart && (headers[keyEnd-1] == ' ' || headers[keyEnd-1] == '\t') {
			keyEnd--
		}

		keyLen := keyEnd - keyStart
		if keyLen != targetLen {
			continue
		}

		// 大小写不敏感比较
		match := true
		for j := 0; j < keyLen; j++ {
			c := headers[keyStart+j]
			if c >= 'A' && c <= 'Z' {
				c = c + ('a' - 'A')
			}
			if c != target[j] {
				match = false
				break
			}
		}

		if !match {
			continue
		}

		// 提取value
		valueStart := colonPos + 1
		valueEnd := lineEnd

		for valueStart < valueEnd && (headers[valueStart] == ' ' || headers[valueStart] == '\t') {
			valueStart++
		}
		for valueEnd > valueStart && (headers[valueEnd-1] == ' ' || headers[valueEnd-1] == '\t') {
			valueEnd--
		}

		if valueEnd > valueStart {
			value := string(headers[valueStart:valueEnd])
			// 提取第一个IP
			if commaPos := strings.IndexByte(value, ','); commaPos != -1 {
				firstIP := strings.TrimSpace(value[:commaPos])
				if firstIP != "" {
					return firstIP, nil
				}
			}
			return strings.TrimSpace(value), nil
		}
		return "", nil
	}

	return "", nil
}

// ==================== 基准测试 ====================

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
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
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
				_, _ = getHeaderValueOptimized(testHeaders, header)
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
			_, _ = getHeaderValueOptimized(testHeaders, "target-header")
		}
	})

	b.Run("Old_Worst_Case", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "target-header")
		}
	})
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

	// 创建模拟的SrcIp
	srcIp, _ := netip.ParseAddr("203.0.113.1")

	for _, tc := range testCases {
		req := &ApplicationRequest{
			Headers: tc.headers,
			SrcIp:   srcIp,
		}

		b.Run(tc.name+"_Original", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIPOriginal(req)
			}
		})

		b.Run(tc.name+"_Optimized", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIPOptimized(req)
			}
		})
	}
}

// 专门测试X-Forwarded-For快速路径的性能
func BenchmarkXForwardedForFastPath(b *testing.B) {
	testHeaders := []byte(`X-Forwarded-For: 192.168.1.100, 10.0.0.1, 172.16.0.1
Host: example.com
User-Agent: Mozilla/5.0`)

	b.Run("FastPath_XFF", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getXForwardedForIP(testHeaders)
		}
	})

	b.Run("Generic_XFF", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
		}
	})
}

// 内存分配测试
func BenchmarkHeaderParsingMemory(b *testing.B) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100, 172.16.0.1
Content-Type: application/json; charset=utf-8`)

	b.Run("Old_Implementation_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "x-forwarded-for")
		}
	})

	b.Run("New_Implementation_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
		}
	})
}
