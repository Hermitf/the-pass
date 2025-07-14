package userutils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT配置结构
type JWTConfig struct {
    SecretKey string
    ExpiresIn int64 // 过期时间（秒）
}

// getJWTSecret 获取JWT密钥
func getJWTSecret() string {
    // 优先从环境变量获取
    if secret := os.Getenv("JWT_SECRET"); secret != "" {
        return secret
    }
    
    // 返回默认密钥（实际项目中应该从配置文件读取）
    return "K7gNU3sdo+OL0wNhqoVWhr3g6s1xYv72ol/pe/Unols="
}

// GenerateJWTToken 生成JWT令牌
func GenerateJWTToken(userID int64) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "user_id": strconv.FormatInt(userID, 10),
        "exp":     now.Add(time.Hour * 24).Unix(),
        "iat":     now.Unix(),
        "nbf":     now.Unix(),                      // Not Before
        "jti":     strconv.FormatInt(now.UnixNano(), 10), // JWT ID (纳秒时间戳确保唯一性)
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err := token.SignedString([]byte(getJWTSecret()))
    if err != nil {
        return "", err
    }
    
    return signedToken, nil
}

// GenerateJWTTokenWithConfig 使用自定义配置生成JWT令牌
func GenerateJWTTokenWithConfig(userID int64, config *JWTConfig) (string, error) {
    if config == nil {
        return "", errors.New("JWT配置不能为空")
    }
    
    now := time.Now()
    claims := jwt.MapClaims{
        "user_id": strconv.FormatInt(userID, 10),
        "exp":     now.Add(time.Duration(config.ExpiresIn) * time.Second).Unix(),
        "iat":     now.Unix(),
        "nbf":     now.Unix(),
        "jti":     strconv.FormatInt(now.UnixNano(), 10), // JWT ID (纳秒时间戳确保唯一性)
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err := token.SignedString([]byte(config.SecretKey))
    if err != nil {
        return "", err
    }
    
    return signedToken, nil
}

// VerifyJWTToken 验证JWT令牌
func VerifyJWTToken(tokenString string) (int64, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // 验证签名方法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return []byte(getJWTSecret()), nil
    })
    
    if err != nil {
        return 0, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if userIDStr, ok := claims["user_id"].(string); ok {
            userID, err := strconv.ParseInt(userIDStr, 10, 64)
            if err != nil {
                return 0, errors.New("无效的用户ID格式")
            }
            return userID, nil
        }
    }
    
    return 0, errors.New("无效的令牌")
}

// VerifyJWTTokenWithConfig 使用自定义配置验证JWT令牌
func VerifyJWTTokenWithConfig(tokenString string, config *JWTConfig) (int64, error) {
    if config == nil {
        return 0, errors.New("JWT配置不能为空")
    }
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return []byte(config.SecretKey), nil
    })
    
    if err != nil {
        return 0, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if userIDStr, ok := claims["user_id"].(string); ok {
            userID, err := strconv.ParseInt(userIDStr, 10, 64)
            if err != nil {
                return 0, errors.New("无效的用户ID格式")
            }
            return userID, nil
        }
    }
    
    return 0, errors.New("无效的令牌")
}

// GetJWTClaims 获取JWT令牌中的所有声明
func GetJWTClaims(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return []byte(getJWTSecret()), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("无效的令牌")
}

// IsJWTTokenExpired 检查JWT令牌是否过期
func IsJWTTokenExpired(tokenString string) bool {
    claims, err := GetJWTClaims(tokenString)
    if err != nil {
        return true
    }
    
    if exp, ok := claims["exp"].(float64); ok {
        return time.Now().Unix() > int64(exp)
    }
    
    return true
}

// RefreshJWTToken 刷新JWT令牌
func RefreshJWTToken(tokenString string) (string, error) {
    userID, err := VerifyJWTToken(tokenString)
    if err != nil {
        return "", err
    }
    
    // 添加微小的时间延迟或使用纳秒级时间戳
    time.Sleep(1 * time.Millisecond)

    return GenerateJWTToken(userID)
}

// IsJWTTokenExpiredWithConfig 使用自定义配置检查JWT令牌是否过期
func IsJWTTokenExpiredWithConfig(tokenString string, config *JWTConfig) bool {
    if config == nil {
        return true
    }
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return []byte(config.SecretKey), nil
    })
    
    if err != nil {
        return true
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if exp, ok := claims["exp"].(float64); ok {
            return time.Now().Unix() > int64(exp)
        }
    }
    
    return true
}

// GetJWTClaimsWithConfig 使用自定义配置获取JWT令牌中的所有声明
func GetJWTClaimsWithConfig(tokenString string, config *JWTConfig) (jwt.MapClaims, error) {
    if config == nil {
        return nil, errors.New("JWT配置不能为空")
    }
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return []byte(config.SecretKey), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("无效的令牌")
}