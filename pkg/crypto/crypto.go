// Package crypto 提供隐私数据 AES 加密存储、密码哈希校验等安全工具。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"activity/config"

	"golang.org/x/crypto/bcrypt"
)

// Encrypt 使用 AES-CBC 加密字符串，返回 Base64 密文。空字符串原样返回。
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	cfg := config.Get().AES
	block, err := aes.NewCipher([]byte(cfg.Key))
	if err != nil {
		return "", fmt.Errorf("AES 密钥非法（需 32 字节）：%w", err)
	}
	raw := []byte(plain)
	raw = padPKCS7(raw, block.BlockSize())
	buf := make([]byte, aes.BlockSize+len(raw))
	iv := buf[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(buf[aes.BlockSize:], raw)
	return base64.StdEncoding.EncodeToString(buf), nil
}

// Decrypt 解密 AES-CBC 密文。空字符串或解密失败时返回原文，避免脏数据导致服务异常。
func Decrypt(cipherText string) string {
	if cipherText == "" {
		return ""
	}
	cfg := config.Get().AES
	block, err := aes.NewCipher([]byte(cfg.Key))
	if err != nil {
		return cipherText
	}
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil || len(raw) <= aes.BlockSize || (len(raw)%aes.BlockSize) != 0 {
		// 兼容历史明文数据
		return cipherText
	}
	buf := make([]byte, len(raw)-aes.BlockSize)
	copy(buf, raw[aes.BlockSize:])
	mode := cipher.NewCBCDecrypter(block, raw[:aes.BlockSize])
	mode.CryptBlocks(buf, buf)
	buf, err = unpadPKCS7(buf)
	if err != nil {
		return cipherText
	}
	return string(buf)
}

// MaskPhone 手机号脱敏：138****8888
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskIDCard 身份证脱敏：前 6 后 4
func MaskIDCard(id string) string {
	if len(id) < 10 {
		return id
	}
	return id[:6] + strings.Repeat("*", len(id)-10) + id[len(id)-4:]
}

// HashPassword 生成 bcrypt 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验密码是否匹配
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(data, pad...)
}

func unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > len(data) {
		return nil, errors.New("invalid padding")
	}
	return data[:len(data)-padding], nil
}
