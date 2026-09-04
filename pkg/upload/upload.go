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

var allowExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}

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
		return nil
	}
}
