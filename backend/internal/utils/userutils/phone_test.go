package userutils

import (
	"testing"
)

// 测试手机号验证 - 增加更多测试用例
func TestIsPhone(t *testing.T) {
    tests := []struct {
        name     string
        phone    string
        expected bool
    }{
        // 有效手机号
        {"valid mobile 13x", "13812345678", true},
        {"valid mobile 14x", "14712345678", true},
        {"valid mobile 15x", "15987654321", true},
        {"valid mobile 16x", "16612345678", true},
        {"valid mobile 17x", "17712345678", true},
        {"valid mobile 18x", "18812345678", true},
        {"valid mobile 19x", "19912345678", true},
        
        // 无效手机号
        {"invalid too short", "1381234567", false},
        {"invalid too long", "138123456789", false},
        {"invalid not start with 1", "23812345678", false},
        {"invalid second digit 0", "10812345678", false},
        {"invalid second digit 1", "11812345678", false},
        {"invalid second digit 2", "12812345678", false},
        {"invalid contains letters", "1381234567a", false},
        {"invalid contains spaces", "138 1234 5678", false},
        {"invalid contains dashes", "138-1234-5678", false},
        {"invalid contains dots", "138.1234.5678", false},
        {"invalid empty", "", false},
        {"invalid only spaces", "   ", false},
        {"invalid special chars", "138#1234#5678", false},
        {"invalid chinese chars", "138一二三四五六七八", false},
        {"invalid all zeros", "00000000000", false},
        {"invalid all same digit", "11111111111", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsPhone(tt.phone)
            if got != tt.expected {
                t.Errorf("IsPhone(%q) = %v; want %v", tt.phone, got, tt.expected)
            }
        })
    }
}

