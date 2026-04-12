package s3

import (
	"mime"
	"strings"
)

var ExtensionMap = map[string]string{
	"audio/mpeg":  ".mp3",
	"audio/wav":   ".wav",
	"audio/webm":  ".weba",
	"audio/x-aac": ".aac",
	"image/jpeg":  ".jpg",
	"image/png":   ".png",
	"image/webp":  ".webp",
	"video/mp4":   ".mp4",
	"video/webm":  ".webm",
}

func DetectExtensions(contentType string) ([]string, error) {
	base, _, _ := strings.Cut(contentType, ";")
	mediatype := strings.TrimSpace(strings.ToLower(base))

	if ext, ok := ExtensionMap[mediatype]; ok {
		return []string{ext}, nil
	}

	return mime.ExtensionsByType(contentType)
}
