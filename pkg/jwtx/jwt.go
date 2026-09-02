// Package jwtx 封装 JWT 生成与解析，区分小程序用户端与 PC 管理端令牌。
package jwtx

import (
	"errors"
	"time"

	"activity/config"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims 小程序用户令牌载荷
type UserClaims struct {
	UserID int64 `json:"uid"`
	Role   int8  `json:"role"`
	jwt.RegisteredClaims
}

// AdminClaims 管理端令牌载荷
type AdminClaims struct {
	AdminID  int64  `json:"aid"`
	Username string `json:"username"`
	Role     int8   `json:"role"` // 1 超级管理员 2 子管理员
	jwt.RegisteredClaims
}

// GenerateUserToken 生成小程序用户令牌
func GenerateUserToken(userID int64, role int8) (string, int64, error) {
	cfg := config.Get().JWT
	expire := time.Duration(cfg.UserExpire) * time.Hour
	if expire <= 0 {
		expire = 720 * time.Hour
	}
	exp := time.Now().Add(expire)
	claims := UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "activity-user",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
	return token, exp.Unix(), err
}

// GenerateAdminToken 生成管理端令牌
func GenerateAdminToken(adminID int64, username string, role int8) (string, int64, error) {
	cfg := config.Get().JWT
	expire := time.Duration(cfg.AdminExpire) * time.Hour
	if expire <= 0 {
		expire = 12 * time.Hour
	}
	exp := time.Now().Add(expire)
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "activity-admin",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
	return token, exp.Unix(), err
}

// ParseUserToken 解析小程序用户令牌
func ParseUserToken(tokenStr string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Get().JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseAdminToken 解析管理端令牌
func ParseAdminToken(tokenStr string) (*AdminClaims, error) {
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Get().JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
