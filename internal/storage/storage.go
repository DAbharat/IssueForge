package storage

import (
	"context"
	"mime/multipart"
)

type UploadResult struct {
	StorageKey   string
	ResourceType string
	MIMEType     string
	Size         int64
}

type Storage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, storageKey string) (*UploadResult, error)
	Delete(ctx context.Context, storageKey, resourceType string) error
}
