package mime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.source.hueristiq.com/http/mime"
)

func TestMIMEString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		m        mime.MIME
		expected string
	}{
		// Application MIME Types
		{"BinaryData", mime.BinaryData, "application/octet-stream"},
		{"JSON", mime.JSON, "application/json"},
		{"JSONLD", mime.JSONLD, "application/ld+json"},
		{"PDF", mime.PDF, "application/pdf"},
		{"XML", mime.XML, "application/xml"},
		{"XHTML", mime.XHTML, "application/xhtml+xml"},
		{"ZIPArchive", mime.ZIPArchive, "application/zip"},
		{"RARArchive", mime.RARArchive, "application/vnd.rar"},
		{"GZipCompressedArchive", mime.GZipCompressedArchive, "application/gzip"},
		{"TARArchive", mime.TARArchive, "application/x-tar"},
		{"SevenZipArchive", mime.SevenZipArchive, "application/x-7z-compressed"},
		{"JavaArchive", mime.JavaArchive, "application/java-archive"},
		{"OGG", mime.OGG, "application/ogg"},
		{"CShellScript", mime.CShellScript, "application/x-csh"},
		{"BourneShellScript", mime.BourneShellScript, "application/x-sh"},
		{"CDAudio", mime.CDAudio, "application/x-cdf"},

		// Audio MIME Types
		{"AACAudio", mime.AACAudio, "audio/aac"},
		{"MP3Audio", mime.MP3Audio, "audio/mpeg"},
		{"OGGAudio", mime.OGGAudio, "audio/ogg"},
		{"OpusAudio", mime.OpusAudio, "audio/opus"},
		{"WAVAudio", mime.WAVAudio, "audio/wav"},
		{"WEBMAudio", mime.WEBMAudio, "audio/webm"},
		{"MIDI", mime.MIDI, "audio/midi"},

		// Video MIME Types
		{"AVIVideo", mime.AVIVideo, "video/x-msvideo"},
		{"MP4Video", mime.MP4Video, "video/mp4"},
		{"MPEGVideo", mime.MPEGVideo, "video/mpeg"},
		{"MPEGTransportStream", mime.MPEGTransportStream, "video/mp2t"},
		{"OGGVideo", mime.OGGVideo, "video/ogg"},
		{"WEBMVideo", mime.WEBMVideo, "video/webm"},
		{"ThreeGPAudioVideo", mime.ThreeGPAudioVideo, "video/3gpp"},
		{"ThreeG2AudioVideo", mime.ThreeG2AudioVideo, "video/3gpp2"},

		// Image MIME Types
		{"AVIFImage", mime.AVIFImage, "image/avif"},
		{"BitmapImage", mime.BitmapImage, "image/bmp"},
		{"GIF", mime.GIF, "image/gif"},
		{"IconFormat", mime.IconFormat, "image/vnd.microsoft.icon"},
		{"JPEG", mime.JPEG, "image/jpeg"},
		{"PNG", mime.PNG, "image/png"},
		{"SVG", mime.SVG, "image/svg+xml"},
		{"TIFF", mime.TIFF, "image/tiff"},
		{"WEBPImage", mime.WEBPImage, "image/webp"},

		// Font MIME Types
		{"OpenTypeFont", mime.OpenTypeFont, "font/otf"},
		{"TrueTypeFont", mime.TrueTypeFont, "font/ttf"},
		{"WOFF", mime.WOFF, "font/woff"},
		{"WOFF2", mime.WOFF2, "font/woff2"},
		{"MSEmbeddedOpenTypeFonts", mime.MSEmbeddedOpenTypeFonts, "application/vnd.ms-fontobject"},

		// Text MIME Types
		{"CSS", mime.CSS, "text/css"},
		{"CSV", mime.CSV, "text/csv"},
		{"HTML", mime.HTML, "text/html"},
		{"ICalendar", mime.ICalendar, "text/calendar"},
		{"JavaScript", mime.JavaScript, "text/javascript"},
		{"PHP", mime.PHP, "application/x-httpd-php"},
		{"RichTextFormat", mime.RichTextFormat, "application/rtf"},
		{"Text", mime.Text, "text/plain"},
		{"JavaScriptModule", mime.JavaScriptModule, "text/javascript"},

		// Document MIME Types
		{"AbiWordDocument", mime.AbiWordDocument, "application/x-abiword"},
		{"AmazonKindleEBook", mime.AmazonKindleEBook, "application/vnd.amazon.ebook"},
		{"AppleInstallerPackage", mime.AppleInstallerPackage, "application/vnd.apple.installer+xml"},
		{"ArchiveDocument", mime.ArchiveDocument, "application/x-freearc"},
		{"BZipArchive", mime.BZipArchive, "application/x-bzip"},
		{"BZip2Archive", mime.BZip2Archive, "application/x-bzip2"},
		{"EPUB", mime.EPUB, "application/epub+zip"},
		{"MSExcel", mime.MSExcel, "application/vnd.ms-excel"},
		{"MSExcelOpenXML", mime.MSExcelOpenXML, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"MSPowerPoint", mime.MSPowerPoint, "application/vnd.ms-powerpoint"},
		{"MSPowerPointOpenXML", mime.MSPowerPointOpenXML, "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"MSVisio", mime.MSVisio, "application/vnd.visio"},
		{"MSWord", mime.MSWord, "application/msword"},
		{"MSWordOpenXML", mime.MSWordOpenXML, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"OpenDocumentPresentation", mime.OpenDocumentPresentation, "application/vnd.oasis.opendocument.presentation"},
		{"OpenDocumentSpreadsheet", mime.OpenDocumentSpreadsheet, "application/vnd.oasis.opendocument.spreadsheet"},
		{"OpenDocumentText", mime.OpenDocumentText, "application/vnd.oasis.opendocument.text"},

		// Other MIME Types
		{"XUL", mime.XUL, "application/vnd.mozilla.xul+xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.m.String()

			assert.Equal(t, tc.expected, actual, "Expected MIME %s to be %s", tc.name, tc.expected)
		})
	}
}

func TestCustomMIME(t *testing.T) {
	t.Parallel()

	custom := mime.MIME("custom/type")

	assert.Equal(t, "custom/type", custom.String(), "Custom MIME type should return its underlying string representation")
}
