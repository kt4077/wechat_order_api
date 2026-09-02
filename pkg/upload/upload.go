// Package upload 统一处理图片上传，支持本地磁盘（local）与七牛云对象存储（qiniu）。
package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"activity/config"
	"activity/pkg/errcode"
	"activity/pkg/logger"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

// allowExt 允许上传的图片扩展名
var allowExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}

// SaveImage 保存单个上传图片并返回可访问 URL。
// 存储驱动由 oss.driver 决定：qiniu 上传至七牛云，local 落盘到 oss.local.root。
func SaveImage(file *multipart.FileHeader) (string, error) {
	cfg := config.Get().OSS
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowExt[ext] {
		return "", errcode.ErrParams.WithMsg("仅支持 jpg/png/gif/webp/bmp 格式图片")
	}
	maxSize := cfg.MaxSize * 1024 * 1024
	if maxSize <= 0 {
		maxSize = 5 * 1024 * 1024
	}
	if file.Size > maxSize {
		return "", errcode.ErrParams.WithMsg(fmt.Sprintf("图片大小不能超过 %dMB", cfg.MaxSize))
	}

	if strings.EqualFold(cfg.Driver, "qiniu") {
		return saveToQiniu(file, ext, cfg)
	}
	return saveToLocal(file, ext, cfg)
}

// saveToLocal 保存到本地磁盘，并通过 /uploads 静态目录对外提供访问
func saveToLocal(file *multipart.FileHeader, ext string, cfg config.OSSConfig) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", errcode.ErrSystem.WithMsg("读取上传文件失败")
	}
	defer src.Close()

	root := cfg.Local.Root
	if root == "" {
		root = "./uploads"
	}
	relDir := time.Now().Format("20060102")
	absDir := filepath.Join(root, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", errcode.ErrSystem.WithMsg("创建上传目录失败")
	}
	relPath := relDir + "/" + fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create(filepath.Join(root, relPath))
	if err != nil {
		return "", errcode.ErrSystem.WithMsg("保存上传文件失败")
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", errcode.ErrSystem.WithMsg("写入上传文件失败")
	}
	if cfg.Local.Domain != "" {
		return strings.TrimRight(cfg.Local.Domain, "/") + "/" + relPath, nil
	}
	return "/uploads/" + relPath, nil
}

// saveToQiniu 使用表单直传方式上传到七牛云对象存储
func saveToQiniu(file *multipart.FileHeader, ext string, cfg config.OSSConfig) (string, error) {
	q := cfg.Qiniu
	if q.Bucket == "" || q.AccessKey == "" || q.SecretKey == "" {
		return "", errcode.ErrSystem.WithMsg("七牛云配置不完整，请检查 oss.qiniu 配置")
	}
	src, err := file.Open()
	if err != nil {
		return "", errcode.ErrSystem.WithMsg("读取上传文件失败")
	}
	defer src.Close()

	mac := auth.New(q.AccessKey, q.SecretKey)
	// 上传凭证默认有效期 1 小时，saveKey 与 key 保持一致
	policy := storage.PutPolicy{Scope: q.Bucket, Expires: 3600}
	upToken := policy.UploadToken(mac)

	prefix := q.Prefix
	if prefix == "" {
		prefix = "activity"
	}
	key := fmt.Sprintf("%s/%s/%d%s", prefix, time.Now().Format("20060102"), time.Now().UnixNano(), ext)

	zone := resolveZone(q.Zone)
	uploader := storage.NewFormUploader(&storage.Config{
		Zone:          zone,
		UseHTTPS:      true,
		UseCdnDomains: false,
	})
	ret := storage.PutRet{}
	if err := uploader.Put(context.Background(), &ret, upToken, key, src, file.Size, &storage.PutExtra{
		TryTimes: 3,
	}); err != nil {
		logger.Errorf("上传七牛云失败：%v", err)
		return "", errcode.ErrSystem.WithMsg("图片上传失败，请稍后重试")
	}
	return strings.TrimRight(q.Domain, "/") + "/" + ret.Key, nil
}

// resolveZone 将区域代号转换为七牛云存储区域配置，留空时由 SDK 自动选择
func resolveZone(code string) *storage.Zone {
	switch strings.ToLower(code) {
	case "z0", "huadong", "east":
		return &storage.ZoneHuadong
	case "z1", "huabei", "north":
		return &storage.ZoneHuabei
	case "z2", "huanan", "south":
		return &storage.ZoneHuanan
	case "na0", "north-america":
		return &storage.ZoneBeimei
	case "as0", "southeast-asia":
		return &storage.ZoneXinjiapo
	default:
		// 未指定区域时由 SDK 依据 bucket 自动查询机房
		return nil
	}
}
