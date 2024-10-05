package headers

// This file list HTTP header field names as constants.
//
// HTTP header fields are components of the header section of request
// and response messages in the HTTP protocol, which is the
// foundation of data communication for the World Wide Web.

const (
	// Authentication - These header fields are related to authentication and authorization.
	Authorization      = "Authorization"
	ProxyAuthenticate  = "Proxy-Authenticate"
	ProxyAuthorization = "Proxy-Authorization"
	WWWAuthenticate    = "WWW-Authenticate"

	// Caching - These header fields are related to caching policy.
	Age           = "Age"
	CacheControl  = "Cache-Control"
	ClearSiteData = "Clear-Site-Data"
	Expires       = "Expires"
	Pragma        = "Pragma"
	Warning       = "Warning"

	// Client hints - These header fields are used for content adaptation.
	AcceptCH         = "Accept-CH"
	AcceptCHLifetime = "Accept-CH-Lifetime"
	ContentDPR       = "Content-DPR"
	DPR              = "DPR"
	EarlyData        = "Early-Data"
	SaveData         = "Save-Data"
	ViewportWidth    = "Viewport-Width"
	Width            = "Width"

	// Conditionals - These header fields are used for conditional requests.
	ETag              = "ETag"
	IfMatch           = "If-Match"
	IfModifiedSince   = "If-Modified-Since"
	IfNoneMatch       = "If-None-Match"
	IfUnmodifiedSince = "If-Unmodified-Since"
	LastModified      = "Last-Modified"
	Vary              = "Vary"

	// Connection management - These header fields are related to connection management.
	Connection      = "Connection"
	KeepAlive       = "Keep-Alive"
	ProxyConnection = "Proxy-Connection"

	// Content negotiation - These header fields are used for content negotiation.
	Accept         = "Accept"
	AcceptCharset  = "Accept-Charset"
	AcceptEncoding = "Accept-Encoding"
	AcceptLanguage = "Accept-Language"

	// Controls - These header fields are related to general controls.
	Cookie      = "Cookie"
	Expect      = "Expect"
	MaxForwards = "Max-Forwards"
	SetCookie   = "Set-Cookie"

	// CORS (Cross-Origin Resource Sharing) - These header fields are related to CORS.
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	AccessControlMaxAge           = "Access-Control-Max-Age"
	AccessControlRequestHeaders   = "Access-Control-Request-Headers"
	AccessControlRequestMethod    = "Access-Control-Request-Method"
	Origin                        = "Origin"
	TimingAllowOrigin             = "Timing-Allow-Origin"
	XPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies"

	// Do Not Track - These header fields are related to user tracking preference.
	DNT = "DNT"
	Tk  = "Tk"

	// Downloads - This header field is related to content disposition.
	ContentDisposition = "Content-Disposition"

	// Message body information - These header fields are related to the message body.
	ContentEncoding = "Content-Encoding"
	ContentLanguage = "Content-Language"
	ContentLength   = "Content-Length"
	ContentLocation = "Content-Location"
	ContentType     = "Content-Type"

	// Proxies - These header fields are related to proxy servers.
	Forwarded       = "Forwarded"
	Via             = "Via"
	XForwardedFor   = "X-Forwarded-For"
	XForwardedHost  = "X-Forwarded-Host"
	XForwardedProto = "X-Forwarded-Proto"

	// Redirects - This header field is related to HTTP redirection.
	Location = "Location"

	// Request context - These header fields are related to the request context.
	From           = "From"
	Host           = "Host"
	Referer        = "Referer"
	ReferrerPolicy = "Referrer-Policy"
	UserAgent      = "User-Agent"

	// Response context - These header fields are related to the response context.
	Allow  = "Allow"
	Server = "Server"

	// Range requests - These header fields are related to partial requests and responses.
	AcceptRanges = "Accept-Ranges"
	ContentRange = "Content-Range"
	IfRange      = "If-Range"
	Range        = "Range"

	// Security - These header fields are related to security.
	ContentSecurityPolicy           = "Content-Security-Policy"
	ContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	CrossOriginResourcePolicy       = "Cross-Origin-Resource-Policy"
	ExpectCT                        = "Expect-CT"
	FeaturePolicy                   = "Feature-Policy"
	PublicKeyPins                   = "Public-Key-Pins"
	PublicKeyPinsReportOnly         = "Public-Key-Pins-Report-Only"
	StrictTransportSecurity         = "Strict-Transport-Security"
	UpgradeInsecureRequests         = "Upgrade-Insecure-Requests"
	XContentTypeOptions             = "X-Content-Type-Options"
	XDownloadOptions                = "X-Download-Options"
	XFrameOptions                   = "X-Frame-Options"
	XPoweredBy                      = "X-Powered-By"
	XXSSProtection                  = "X-XSS-Protection"

	// Server-sent event - These header fields are related to server-sent events.
	LastEventID = "Last-Event-ID"
	NEL         = "NEL"
	PingFrom    = "Ping-From"
	PingTo      = "Ping-To"
	ReportTo    = "Report-To"

	// Transfer coding - These header fields are related to transfer coding.
	TE               = "TE"
	Trailer          = "Trailer"
	TransferEncoding = "Transfer-Encoding"

	// WebSockets -These header fields are related to the WebSocket protocol.
	SecWebSocketAccept     = "Sec-WebSocket-Accept"
	SecWebSocketExtensions = "Sec-WebSocket-Extensions"
	SecWebSocketKey        = "Sec-WebSocket-Key"
	SecWebSocketProtocol   = "Sec-WebSocket-Protocol"
	SecWebSocketVersion    = "Sec-WebSocket-Version"

	// Other -These header fields do not fall into the above categories.
	AcceptPatch         = "Accept-Patch"
	AcceptPushPolicy    = "Accept-Push-Policy"
	AcceptSignature     = "Accept-Signature"
	AltSvc              = "Alt-Svc"
	Date                = "Date"
	Index               = "Index"
	LargeAllocation     = "Large-Allocation"
	Link                = "Link"
	PushPolicy          = "Push-Policy"
	RetryAfter          = "Retry-After"
	XRatelimitRemaining = "X-Ratelimit-Remaining"
	ServerTiming        = "Server-Timing"
	Signature           = "Signature"
	SignedHeaders       = "Signed-Headers"
	SourceMap           = "SourceMap"
	Upgrade             = "Upgrade"
	XDNSPrefetchControl = "X-DNS-Prefetch-Control"
	XPingback           = "X-Pingback"
	XRequestedWith      = "X-Requested-With"
	XRobotsTag          = "X-Robots-Tag"
	XUACompatible       = "X-UA-Compatible"
)
