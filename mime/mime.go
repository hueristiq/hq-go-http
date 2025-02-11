package mime

// MIME represents Internet Media Types as defined by IANA.
//
// The MIME type provides a strongly typed representation of media types (or content types),
// ensuring correctness and maintainability when handling HTTP requests and responses.
//
// Reference: https://www.iana.org/assignments/media-types/media-types.xhtml
type MIME string

// String returns the string representation of the MIME type.
//
// It converts the MIME value (a string alias) to its underlying string type,
// making it suitable for use in HTTP headers or any other context where a plain string is required.
//
// Returns:
//   - mime (string): The MIME type as a string value.
//
// Example:
//
//	m := JSON
//	fmt.Println(m.String()) // Output: "application/json"
func (m MIME) String() (mime string) {
	return string(m)
}

// Interface defines a common interface for MIME types.
//
// Any type that implements a String() method returning a string satisfies this interface.
// This allows for different implementations of MIME types to be used interchangeably.
type Interface interface {
	String() (mime string)
}

// Application MIME Types: Used for non-text and binary data.
const (
	BinaryData            MIME = "application/octet-stream"    // Generic binary data
	JSON                  MIME = "application/json"            // JavaScript Object Notation
	JSONLD                MIME = "application/ld+json"         // JSON for Linked Data
	PDF                   MIME = "application/pdf"             // Portable Document Format
	XML                   MIME = "application/xml"             // Extensible Markup Language
	XHTML                 MIME = "application/xhtml+xml"       // XHTML documents
	ZIPArchive            MIME = "application/zip"             // ZIP archive
	RARArchive            MIME = "application/vnd.rar"         // RAR archive
	GZipCompressedArchive MIME = "application/gzip"            // GZip compressed archive
	TARArchive            MIME = "application/x-tar"           // TAR archive
	SevenZipArchive       MIME = "application/x-7z-compressed" // 7-Zip archive
	JavaArchive           MIME = "application/java-archive"    // Java archive (JAR)
	OGG                   MIME = "application/ogg"             // OGG container format
	CShellScript          MIME = "application/x-csh"           // C Shell script
	BourneShellScript     MIME = "application/x-sh"            // Bourne shell script
	CDAudio               MIME = "application/x-cdf"           // Audio data in CDF format
)

// Audio MIME Types: Used for representing audio files.
const (
	AACAudio  MIME = "audio/aac"  // Advanced Audio Coding
	MP3Audio  MIME = "audio/mpeg" // MPEG audio format
	OGGAudio  MIME = "audio/ogg"  // OGG audio format
	OpusAudio MIME = "audio/opus" // Opus audio codec
	WAVAudio  MIME = "audio/wav"  // Waveform Audio File Format
	WEBMAudio MIME = "audio/webm" // WEBM audio container
	MIDI      MIME = "audio/midi" // Musical Instrument Digital Interface
)

// Video MIME Types: Used for representing video files.
const (
	AVIVideo            MIME = "video/x-msvideo" // AVI video format
	MP4Video            MIME = "video/mp4"       // MP4 video format
	MPEGVideo           MIME = "video/mpeg"      // MPEG video format
	MPEGTransportStream MIME = "video/mp2t"      // MPEG transport stream
	OGGVideo            MIME = "video/ogg"       // OGG video format
	WEBMVideo           MIME = "video/webm"      // WEBM video format
	ThreeGPAudioVideo   MIME = "video/3gpp"      // 3GPP audio/video container
	ThreeG2AudioVideo   MIME = "video/3gpp2"     // 3GPP2 audio/video container
)

// Image MIME Types: Used for representing image files.
const (
	AVIFImage   MIME = "image/avif"               // AVIF image format
	BitmapImage MIME = "image/bmp"                // Bitmap image format
	GIF         MIME = "image/gif"                // Graphics Interchange Format image
	IconFormat  MIME = "image/vnd.microsoft.icon" // Microsoft icon image format
	JPEG        MIME = "image/jpeg"               // JPEG image format
	PNG         MIME = "image/png"                // Portable Network Graphics image format
	SVG         MIME = "image/svg+xml"            // Scalable Vector Graphics
	TIFF        MIME = "image/tiff"               // Tagged Image File Format
	WEBPImage   MIME = "image/webp"               // WEBP image format
)

// Font MIME Types: Used for representing font files.
const (
	OpenTypeFont            MIME = "font/otf"                      // OpenType Font
	TrueTypeFont            MIME = "font/ttf"                      // TrueType Font
	WOFF                    MIME = "font/woff"                     // Web Open Font Format
	WOFF2                   MIME = "font/woff2"                    // Web Open Font Format 2
	MSEmbeddedOpenTypeFonts MIME = "application/vnd.ms-fontobject" // Microsoft Embedded OpenType fonts
)

// Text MIME Types: Used for text-based content.
const (
	CSS              MIME = "text/css"                // Cascading Style Sheets
	CSV              MIME = "text/csv"                // Comma-Separated Values
	HTML             MIME = "text/html"               // HyperText Markup Language
	ICalendar        MIME = "text/calendar"           // Calendar format (iCalendar)
	JavaScript       MIME = "text/javascript"         // JavaScript code
	PHP              MIME = "application/x-httpd-php" // PHP script
	RichTextFormat   MIME = "application/rtf"         // Rich Text Format
	Text             MIME = "text/plain"              // Plain text
	JavaScriptModule MIME = "text/javascript"         // JavaScript module (alias to JavaScript)
)

// Document MIME Types: Used for various document formats.
const (
	AbiWordDocument          MIME = "application/x-abiword"                                                     // AbiWord document
	AmazonKindleEBook        MIME = "application/vnd.amazon.ebook"                                              // Amazon Kindle eBook format
	AppleInstallerPackage    MIME = "application/vnd.apple.installer+xml"                                       // Apple installer package
	ArchiveDocument          MIME = "application/x-freearc"                                                     // Freearc archive
	BZipArchive              MIME = "application/x-bzip"                                                        // BZip archive
	BZip2Archive             MIME = "application/x-bzip2"                                                       // BZip2 archive
	EPUB                     MIME = "application/epub+zip"                                                      // EPUB eBook
	MSExcel                  MIME = "application/vnd.ms-excel"                                                  // Microsoft Excel document
	MSExcelOpenXML           MIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"         // Microsoft Excel (OpenXML)
	MSPowerPoint             MIME = "application/vnd.ms-powerpoint"                                             // Microsoft PowerPoint presentation
	MSPowerPointOpenXML      MIME = "application/vnd.openxmlformats-officedocument.presentationml.presentation" // Microsoft PowerPoint (OpenXML)
	MSVisio                  MIME = "application/vnd.visio"                                                     // Microsoft Visio diagram
	MSWord                   MIME = "application/msword"                                                        // Microsoft Word document
	MSWordOpenXML            MIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"   // Microsoft Word (OpenXML)
	OpenDocumentPresentation MIME = "application/vnd.oasis.opendocument.presentation"                           // OpenDocument presentation
	OpenDocumentSpreadsheet  MIME = "application/vnd.oasis.opendocument.spreadsheet"                            // OpenDocument spreadsheet
	OpenDocumentText         MIME = "application/vnd.oasis.opendocument.text"                                   // OpenDocument text document
)

// Other MIME Types: Used for additional data types.
const (
	XUL MIME = "application/vnd.mozilla.xul+xml" // XUL (XML User Interface Language) document
)

// This compile-time assertion ensures that the MIME type correctly implements the Interface interface.
// If it does not, the assignment will cause a compile-time error.
var _ Interface = (*MIME)(nil)
