package formatting

import (
	"strings"
)

// MaskEmail 遮罩邮箱地址
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	atIndex := strings.Index(email, "@")
	if atIndex < 0 {
		return email
	}

	localPart := email[:atIndex]
	domainPart := email[atIndex:]

	if len(localPart) <= 2 {
		return "***" + domainPart
	}

	if len(localPart) <= 4 {
		return localPart[:1] + "***" + domainPart
	}

	return localPart[:2] + "***" + domainPart
}

// MaskPhone 遮罩手机号
func MaskPhone(phone string) string {
	if phone == "" {
		return ""
	}

	// 清理手机号
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")

	if len(cleanPhone) != 11 || !strings.HasPrefix(cleanPhone, "1") {
		return phone
	}

	return cleanPhone[:3] + "****" + cleanPhone[7:]
}
