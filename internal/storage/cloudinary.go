package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cld *cloudinary.Cloudinary) *CloudinaryStorage {
	return &CloudinaryStorage{
		client: cld,
	}
}

var (
	ErrEmptyFile       = errors.New("file is empty")
	ErrFileTooLarge    = errors.New("file exceeds maximum allowed size of 10MB")
	ErrUnsupportedType = errors.New("unsupported file content type")
	ErrDeleteFailed    = errors.New("failed to delete file from storage")
)

const maxFileSize = 10 << 20

var allowedMIMETypes = map[string]string{
	"image/jpeg":      "image",
	"image/png":       "image",
	"image/webp":      "image",
	"application/pdf": "raw",
}

func (s *CloudinaryStorage) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, storageKey string) (*UploadResult, error) {
	if header.Size <= 0 {
		return nil, ErrEmptyFile
	}
	if header.Size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}

	mimeType := http.DetectContentType(buffer[:n])

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("seek file start: %w", err)
	}

	resourceType, ok := allowedMIMETypes[mimeType]
	if !ok {
		return nil, ErrUnsupportedType
	}

	overwrite := false
	unique := false

	resp, err := s.client.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       storageKey,
		ResourceType:   resourceType,
		Overwrite:      &overwrite,
		UniqueFilename: &unique,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload: %w", err)
	}

	return &UploadResult{
		StorageKey:   resp.PublicID,
		ResourceType: resourceType,
		MIMEType:     mimeType,
		Size:         header.Size,
	}, nil
}

func (s *CloudinaryStorage) Delete(ctx context.Context, storageKey, resourceType string) error {
	invalidate := true
	params := uploader.DestroyParams{
		PublicID:     storageKey,
		ResourceType: resourceType,
		Invalidate:   &invalidate,
	}

	resp, err := s.client.Upload.Destroy(ctx, params)
	if err != nil {
		return fmt.Errorf("cloudinary destroy: %w", err)
	}

	if resp.Result != "ok" {
		return fmt.Errorf("%w: status %s", ErrDeleteFailed, resp.Result)
	}
	return nil
}
