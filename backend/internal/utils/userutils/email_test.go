package userutils

import (
	"strings"
	"testing"
)

// 测试邮箱验证 - 增加更多测试用例
func TestIsEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		// 有效邮箱
		{"valid basic email", "test@example.com", true},
		{"valid with subdomain", "user@mail.google.com", true},
		{"valid with numbers", "user123@test123.com", true},
		{"valid with dots", "first.last@example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with hyphen", "user-name@example.com", true},
		{"valid with underscore", "user_name@example.com", true},
		{"valid short domain", "test@a.co", true},
		{"valid long domain", "test@verylongdomainname.com", true},

		// 无效邮箱
		{"invalid no @", "testexample.com", false},
		{"invalid multiple @", "test@@example.com", false},
		{"invalid missing username", "@example.com", false},
		{"invalid missing domain", "test@", false},
		{"invalid missing tld", "test@example", false},
		{"invalid spaces", "test @example.com", false},
		{"invalid special chars", "test$@example.com", false},
		{"invalid consecutive dots", "test..user@example.com", false},
		{"invalid dot at start", ".test@example.com", false},
		{"invalid dot at end", "test.@example.com", false},
		{"invalid empty", "", false},
		{"invalid only spaces", "   ", false},
		{"invalid chinese chars", "测试@example.com", false},
		{"invalid emoji", "test😀@example.com", false},
		{"invalid too long", strings.Repeat("a", 250) + "@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmail(tt.email)
			if got != tt.expected {
				t.Errorf("IsEmail(%q) = %v; want %v", tt.email, got, tt.expected)
			}
		})
	}
}
