package internal

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// 原始的bufio.Scanner实现（用于对比验证）
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

// TestGetHeaderValue 测试getHeaderValue函数的正确性
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

			// 测试与旧实现的一致性
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

// TestGetHeaderValueEdgeCases 测试边界情况
func TestGetHeaderValueEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		headers  []byte
		header   string
		expected string
	}{
		{
			name:     "空headers",
			headers:  []byte(""),
			header:   "test",
			expected: "",
		},
		{
			name:     "只有换行符",
			headers:  []byte("\n\n\n"),
			header:   "test",
			expected: "",
		},
		{
			name:     "无冒号的行",
			headers:  []byte("invalid line\nHost: example.com"),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "value前后有空格",
			headers:  []byte("Host:   example.com   "),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "key前后有空格",
			headers:  []byte("   Host   : example.com"),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "空value",
			headers:  []byte("Empty-Header: "),
			header:   "empty-header",
			expected: "",
		},
		{
			name:     "多个冒号",
			headers:  []byte("Time: 12:34:56"),
			header:   "time",
			expected: "12:34:56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getHeaderValue(tt.headers, tt.header)
			if err != nil {
				t.Errorf("getHeaderValue() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("getHeaderValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGetRealClientIP 测试getRealClientIP函数的正确性
func TestGetRealClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  []byte
		expected string
	}{
		{
			name: "X-Forwarded-For单个IP",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100`),
			expected: "192.168.1.100",
		},
		{
			name: "X-Forwarded-For多个IP",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100, 10.0.0.1, 172.16.0.1`),
			expected: "192.168.1.100",
		},
		{
			name: "X-Real-IP",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.200`),
			expected: "192.168.1.200",
		},
		{
			name: "CF-Connecting-IP",
			headers: []byte(`Host: example.com
CF-Connecting-IP: 1.2.3.4`),
			expected: "1.2.3.4",
		},
		{
			name: "优先级测试：X-Forwarded-For优先",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.200
X-Forwarded-For: 192.168.1.100`),
			expected: "192.168.1.100",
		},
		{
			name: "Forwarded标准头部",
			headers: []byte(`Host: example.com
Forwarded: for=192.168.1.100;proto=https;by=proxy`),
			expected: "192.168.1.100",
		},
		{
			name:     "无客户端IP头部",
			headers:  []byte(`Host: example.com\nUser-Agent: test`),
			expected: "", // 由于没有设置SrcIp，应该返回空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &applicationRequest{
				Headers: tt.headers,
			}
			result := getRealClientIP(req)
			if result != tt.expected {
				t.Errorf("getRealClientIP() = %q, want %q", result, tt.expected)
			}
		})
	}
}
