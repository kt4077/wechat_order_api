package jwtx

import (
	"errors"
	"time"

	"activity/config"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID int64 `json:"uid"`
	Role   int8  `json:"role"`
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminID  int64  `json:"aid"`
	Username string `json:"username"`
	Role     int8   `json:"role"`
	jwt.RegisteredClaims
}

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
