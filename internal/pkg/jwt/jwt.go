package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// 生成访问令牌
func GenerateToken(secret string, userID int64, expireDuration time.Duration) (string, error) {
	now := time.Now()
	jwtClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "user-service",
			Subject:   "access-token",
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(secret))
}

// 解析令牌
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// 生成刷新令牌
func GenerateRefreshToken(secret string, userID int64, expireDuration time.Duration) (string, error) {
	now := time.Now()
	jwtClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "user-service",
			Subject:   "refresh-token",
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(secret))
}
