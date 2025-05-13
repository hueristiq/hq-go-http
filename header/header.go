package header

// Header represents an HTTP header field as defined by IANA.
//
// The underlying type is string, which allows seamless conversion and integration
// with other string-based operations. Using a custom type like Header helps to ensure
// that only valid and predefined HTTP header names are used throughout the codebase.
type Header string

// String converts the Header to its underlying string representation.
//
// This method is used to retrieve the actual header name as a string, which is useful
// when interacting with other parts of an HTTP framework or when outputting header values.
//
// Returns:
//   - header (string): The HTTP header field as a string.
func (h Header) String() (header string) {
	header = string(h)

	return
}

// Predefined HTTP header constants.
//
// The constants below represent commonly used HTTP header fields as defined by IANA.
// They are declared as type Header to ensure type safety and to minimize the possibility
// of typos when using header names throughout the code.
//
// These constants can be referenced directly to improve code clarity and maintainability.
const (
	AccessControlAllowCredentials   Header = "Access-Control-Allow-Credentials"
	AccessControlAllowHeaders       Header = "Access-Control-Allow-Headers"
	AccessControlAllowMethods       Header = "Access-Control-Allow-Methods"
	AccessControlAllowOrigin        Header = "Access-Control-Allow-Origin"
	AccessControlExposeHeaders      Header = "Access-Control-Expose-Headers"
	AccessControlMaxAge             Header = "Access-Control-Max-Age"
	AccessControlRequestHeaders     Header = "Access-Control-Request-Headers"
	AccessControlRequestMethod      Header = "Access-Control-Request-Method"
	Accept                          Header = "Accept"
	AcceptCH                        Header = "Accept-CH"
	AcceptCHLifetime                Header = "Accept-CH-Lifetime"
	AcceptCharset                   Header = "Accept-Charset"
	AcceptEncoding                  Header = "Accept-Encoding"
	AcceptLanguage                  Header = "Accept-Language"
	AcceptPatch                     Header = "Accept-Patch"
	AcceptPushPolicy                Header = "Accept-Push-Policy"
	AcceptRanges                    Header = "Accept-Ranges"
	AcceptSignature                 Header = "Accept-Signature"
	Age                             Header = "Age"
	Allow                           Header = "Allow"
	AltSvc                          Header = "Alt-Svc"
	Authorization                   Header = "Authorization"
	CacheControl                    Header = "Cache-Control"
	ClearSiteData                   Header = "Clear-Site-Data"
	Connection                      Header = "Connection"
	ContentDPR                      Header = "Content-DPR"
	ContentDisposition              Header = "Content-Disposition"
	ContentEncoding                 Header = "Content-Encoding"
	ContentLanguage                 Header = "Content-Language"
	ContentLength                   Header = "Content-Length"
	ContentLocation                 Header = "Content-Location"
	ContentRange                    Header = "Content-Range"
	ContentSecurityPolicy           Header = "Content-Security-Policy"
	ContentSecurityPolicyReportOnly Header = "Content-Security-Policy-Report-Only"
	ContentType                     Header = "Content-Type"
	Cookie                          Header = "Cookie"
	CrossOriginResourcePolicy       Header = "Cross-Origin-Resource-Policy"
	DPR                             Header = "DPR"
	DNT                             Header = "DNT"
	Date                            Header = "Date"
	EarlyData                       Header = "Early-Data"
	ETag                            Header = "ETag"
	Expect                          Header = "Expect"
	ExpectCT                        Header = "Expect-CT"
	Expires                         Header = "Expires"
	FeaturePolicy                   Header = "Feature-Policy"
	Forwarded                       Header = "Forwarded"
	From                            Header = "From"
	Host                            Header = "Host"
	IfMatch                         Header = "If-Match"
	IfModifiedSince                 Header = "If-Modified-Since"
	IfNoneMatch                     Header = "If-None-Match"
	IfRange                         Header = "If-Range"
	IfUnmodifiedSince               Header = "If-Unmodified-Since"
	Index                           Header = "Index"
	KeepAlive                       Header = "Keep-Alive"
	LargeAllocation                 Header = "Large-Allocation"
	LastEventID                     Header = "Last-Event-ID"
	LastModified                    Header = "Last-Modified"
	Link                            Header = "Link"
	Location                        Header = "Location"
	MaxForwards                     Header = "Max-Forwards"
	NEL                             Header = "NEL"
	Origin                          Header = "Origin"
	PingFrom                        Header = "Ping-From"
	PingTo                          Header = "Ping-To"
	Pragma                          Header = "Pragma"
	ProxyAuthenticate               Header = "Proxy-Authenticate"
	ProxyAuthorization              Header = "Proxy-Authorization"
	ProxyConnection                 Header = "Proxy-Connection"
	PushPolicy                      Header = "Push-Policy"
	Range                           Header = "Range"
	Referer                         Header = "Referer"
	ReferrerPolicy                  Header = "Referrer-Policy"
	ReportTo                        Header = "Report-To"
	RetryAfter                      Header = "Retry-After"
	SaveData                        Header = "Save-Data"
	SecWebSocketAccept              Header = "Sec-WebSocket-Accept"
	SecWebSocketExtensions          Header = "Sec-WebSocket-Extensions"
	SecWebSocketKey                 Header = "Sec-WebSocket-Key"
	SecWebSocketProtocol            Header = "Sec-WebSocket-Protocol"
	SecWebSocketVersion             Header = "Sec-WebSocket-Version"
	Server                          Header = "Server"
	ServerTiming                    Header = "Server-Timing"
	SetCookie                       Header = "Set-Cookie"
	Signature                       Header = "Signature"
	SignedHeaders                   Header = "Signed-Headers"
	SourceMap                       Header = "SourceMap"
	StrictTransportSecurity         Header = "Strict-Transport-Security"
	TE                              Header = "TE"
	TimingAllowOrigin               Header = "Timing-Allow-Origin"
	Tk                              Header = "Tk"
	Trailer                         Header = "Trailer"
	TransferEncoding                Header = "Transfer-Encoding"
	Upgrade                         Header = "Upgrade"
	UpgradeInsecureRequests         Header = "Upgrade-Insecure-Requests"
	UserAgent                       Header = "User-Agent"
	Vary                            Header = "Vary"
	Via                             Header = "Via"
	ViewportWidth                   Header = "Viewport-Width"
	Warning                         Header = "Warning"
	WWWAuthenticate                 Header = "WWW-Authenticate"
	Width                           Header = "Width"
	XContentTypeOptions             Header = "X-Content-Type-Options"
	XDNSPrefetchControl             Header = "X-DNS-Prefetch-Control"
	XDownloadOptions                Header = "X-Download-Options"
	XFrameOptions                   Header = "X-Frame-Options"
	XForwardedFor                   Header = "X-Forwarded-For"
	XForwardedHost                  Header = "X-Forwarded-Host"
	XForwardedProto                 Header = "X-Forwarded-Proto"
	XPingback                       Header = "X-Pingback"
	XPermittedCrossDomainPolicies   Header = "X-Permitted-Cross-Domain-Policies"
	XPoweredBy                      Header = "X-Powered-By"
	XRequestedWith                  Header = "X-Requested-With"
	XRobotsTag                      Header = "X-Robots-Tag"
	XUACompatible                   Header = "X-UA-Compatible"
	XXSSProtection                  Header = "X-XSS-Protection"
	XRatelimitRemaining             Header = "X-Ratelimit-Remaining"
)
