package s3

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service interface {
	UploadFromURL(ctx context.Context, srcURL string, filePath string) (string, error)
	UploadBase64Data(ctx context.Context, base64Data string, filePath string) (string, error)
	UploadFile(ctx context.Context, filePath string, contentType string, fileContent io.Reader) (string, error)
	UploadFileJSON(ctx context.Context, filePath string, jsonString string) (string, error)
	GetURL(ctx context.Context, filePath string) string
}

type service struct {
	cfg Config
}

func NewService(cfg Config) Service {
	return &service{
		cfg: cfg,
	}
}

func (s *service) GetURL(_ context.Context, filePath string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.CdnURL, "/"), strings.TrimLeft(filePath, "/"))
}

func (s *service) UploadFile(ctx context.Context, filePath string, contentType string, fileContent io.Reader) (string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return "", err
	}
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.cfg.Bucket,
		Key:         &filePath,
		Body:        fileContent,
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}
	return s.GetURL(ctx, filePath), nil
}

func (s *service) UploadFileJSON(ctx context.Context, filePath string, jsonString string) (string, error) {
	return s.UploadFile(ctx, filePath, "application/json", strings.NewReader(jsonString))
}

func (s *service) UploadFromURL(ctx context.Context, srcURL string, filePath string) (string, error) {
	if strings.HasPrefix(srcURL, "data:") {
		return s.UploadBase64Data(ctx, srcURL, filePath)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	exts, _ := DetectExtensions(contentType)
	if len(exts) > 0 {
		filePath += exts[0]
	}

	clh := resp.Header.Get("Content-Length")
	if clh == "" {
		return "", fmt.Errorf("content length is empty")
	}
	contentLength, _ := strconv.Atoi(clh)
	return s.UploadFile(ctx, filePath, contentType, &readerWithLength{
		Reader: resp.Body,
		length: contentLength,
	})
}

func (s *service) UploadBase64Data(ctx context.Context, base64Data string, filePath string) (string, error) {
	parts := strings.Split(base64Data, ";base64,")
	contentType := strings.Replace(parts[0], "data:", "", 1)
	exts, _ := DetectExtensions(contentType)
	if len(exts) > 0 {
		filePath += exts[0]
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	return s.UploadFile(ctx, filePath, contentType, bytes.NewReader(data))
}

func (s *service) getClient(ctx context.Context) (*s3.Client, error) {
	c, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.cfg.AccessKeyID, s.cfg.SecretAccessKey, "")),
		config.WithRegion(s.cfg.Region),
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
		config.WithBaseEndpoint(s.cfg.Endpoint),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(c, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}
