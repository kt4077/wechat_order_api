package controller

import (
	"activity/pkg/upload"

	"mime/multipart"
)

func uploadImageFile(file *multipart.FileHeader) (string, error) {
	return upload.SaveImage(file)
}
