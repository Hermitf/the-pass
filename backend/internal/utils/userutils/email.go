package userutils

import (
	"regexp"
	"strings"
)

// 邮箱验证主函数
func IsEmail(email string) bool {
    return isValidEmailLength(email) && 
           matchesEmailPattern(email) && 
           isValidEmailStructure(email)
}

// 验证邮箱长度
func isValidEmailLength(email string) bool {
    return len(email) <= 254 && len(email) > 0
}

// 验证邮箱正则表达式
func matchesEmailPattern(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// 验证邮箱结构
func isValidEmailStructure(email string) bool {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return false
    }
    
    local, domain := parts[0], parts[1]
    return isValidLocalPart(local) && isValidDomainPart(domain)
}

// 验证邮箱本地部分
func isValidLocalPart(local string) bool {
    if len(local) == 0 {
        return false
    }
    
    if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
        return false
    }
    
    return !strings.Contains(local, "..")
}

// 验证邮箱域名部分
func isValidDomainPart(domain string) bool {
    if len(domain) == 0 {
        return false
    }
    
    if strings.Contains(domain, "..") {
        return false
    }
    
    return !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}