package mime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
)

func TestMIME_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mime     hqgohttpmime.MIME
		expected string
	}{
		{"HTML", hqgohttpmime.HTML, "text/html"},
		{"JSON", hqgohttpmime.JSON, "application/json"},
		{"JSONLD", hqgohttpmime.JSONLD, "application/ld+json"},
		{"XML", hqgohttpmime.XML, "application/xml"},
		{"Text", hqgohttpmime.Text, "text/plain"},
		{"CSV", hqgohttpmime.CSV, "text/csv"},
		{"PNG", hqgohttpmime.PNG, "image/png"},
		{"JPEG", hqgohttpmime.JPEG, "image/jpeg"},
		{"GIF", hqgohttpmime.GIF, "image/gif"},
		{"PDF", hqgohttpmime.PDF, "application/pdf"},
		{"ZIPArchive", hqgohttpmime.ZIPArchive, "application/zip"},
		{"GZipCompressedArchive", hqgohttpmime.GZipCompressedArchive, "application/gzip"},
		{"MP4Video", hqgohttpmime.MP4Video, "video/mp4"},
		{"MP3Audio", hqgohttpmime.MP3Audio, "audio/mpeg"},
		{"WOFF2", hqgohttpmime.WOFF2, "font/woff2"},
		{"BinaryData", hqgohttpmime.BinaryData, "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.mime.String())
		})
	}
}
