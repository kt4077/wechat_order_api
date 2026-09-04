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

func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func MaskIDCard(id string) string {
	if len(id) < 10 {
		return id
	}
	return id[:6] + strings.Repeat("*", len(id)-10) + id[len(id)-4:]
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

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
