package controller

import (
	"activity/pkg/upload"

	"mime/multipart"
)

// uploadImageFile 统一入口封装上传逻辑，便于后续替换为对象存储
func uploadImageFile(file *multipart.FileHeader) (string, error) {
	return upload.SaveImage(file)
}
