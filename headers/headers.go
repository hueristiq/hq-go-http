package headers

const (
	// Authentication - These header fields are used for authentication and authorization.
	// They are commonly found in request messages where the client needs to authenticate with the server.
	Authorization      = "Authorization"       // Used to pass authentication credentials in the form of bearer tokens or other mechanisms.
	ProxyAuthenticate  = "Proxy-Authenticate"  // Used in responses to indicate that the client must authenticate with the proxy.
	ProxyAuthorization = "Proxy-Authorization" // Used in requests to provide proxy authentication credentials.
	WWWAuthenticate    = "WWW-Authenticate"    // Sent by the server in responses, indicating the required authentication method.

	// Caching - These header fields define caching behavior between client and server.
	// They help manage cacheable resources and specify how caching should be handled.
	Age           = "Age"             // Represents the age of a cached response in seconds.
	CacheControl  = "Cache-Control"   // Controls caching policies such as expiration and revalidation.
	ClearSiteData = "Clear-Site-Data" // Used to instruct the client to clear its cached data.
	Expires       = "Expires"         // Specifies the date/time after which the response is considered stale.
	Pragma        = "Pragma"          // Includes implementation-specific directives that might apply to caching.
	Warning       = "Warning"         // Additional information about cache-related operations.

	// Client hints - These headers provide the server with hints about the client's device or preferences,
	// enabling the server to adapt its response content accordingly.
	AcceptCH         = "Accept-CH"          // Indicates client hints the server supports.
	AcceptCHLifetime = "Accept-CH-Lifetime" // Specifies how long the client should persist client hint preferences.
	ContentDPR       = "Content-DPR"        // Specifies the device pixel ratio.
	DPR              = "DPR"                // Provides the device pixel ratio.
	EarlyData        = "Early-Data"         // Indicates that the request is using early data (e.g., from TLS 1.3 0-RTT).
	SaveData         = "Save-Data"          // Informs the server that the client prefers to conserve data usage.
	ViewportWidth    = "Viewport-Width"     // Provides the width of the layout viewport.
	Width            = "Width"              // Describes the display width of the client's device.

	// Conditionals - These headers are used in conditional requests, allowing the client to
	// make requests that depend on specific conditions, such as resource modification times.
	ETag              = "ETag"                // Identifier for a specific version of a resource, used for conditional requests.
	IfMatch           = "If-Match"            // Executes the request only if the entity matches the given ETag.
	IfModifiedSince   = "If-Modified-Since"   // Performs the request only if the resource has been modified after a specified date.
	IfNoneMatch       = "If-None-Match"       // Executes the request only if the entity does not match the given ETag.
	IfUnmodifiedSince = "If-Unmodified-Since" // Executes the request only if the resource has not been modified since the given date.
	LastModified      = "Last-Modified"       // Indicates the last time the resource was modified.
	Vary              = "Vary"                // Determines how responses can vary based on the request header fields.

	// Connection management - These headers are related to connection management, allowing clients and servers
	// to negotiate how the connection should be handled.
	Connection      = "Connection"       // Controls options for the current connection (e.g., keep-alive).
	KeepAlive       = "Keep-Alive"       // Specifies parameters for keeping the connection alive.
	ProxyConnection = "Proxy-Connection" // Manages the connection behavior when communicating with a proxy.

	// Content negotiation - These headers are used by the client to specify the preferred content formats or encodings.
	Accept         = "Accept"          // Informs the server of the acceptable media types for the response.
	AcceptCharset  = "Accept-Charset"  // Specifies the character sets that are acceptable for the response.
	AcceptEncoding = "Accept-Encoding" // Indicates which content encodings (e.g., gzip, deflate) are acceptable.
	AcceptLanguage = "Accept-Language" // Informs the server of the preferred language for the response.

	// Controls - General control-related headers used for managing specific behaviors in requests and responses.
	Cookie      = "Cookie"       // Used by the client to send stored cookies to the server.
	Expect      = "Expect"       // Informs the server of certain expectations for the request.
	MaxForwards = "Max-Forwards" // Limits the number of times a request can be forwarded by proxies.
	SetCookie   = "Set-Cookie"   // Used by the server to send cookies to the client for storage.

	// CORS (Cross-Origin Resource Sharing) - These headers are related to CORS, which allows or restricts
	// access to resources from different origins.
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"  // Specifies whether credentials are allowed for cross-origin requests.
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"      // Indicates which HTTP headers can be used in cross-origin requests.
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"      // Specifies the allowed methods for cross-origin requests.
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"       // Defines which origins are permitted to access resources.
	AccessControlExposeHeaders    = "Access-Control-Expose-Headers"     // Specifies which headers are exposed to the client.
	AccessControlMaxAge           = "Access-Control-Max-Age"            // Specifies how long the results of a preflight request can be cached.
	AccessControlRequestHeaders   = "Access-Control-Request-Headers"    // Indicates which headers the client will send in the request.
	AccessControlRequestMethod    = "Access-Control-Request-Method"     // Specifies the HTTP method the client intends to use.
	Origin                        = "Origin"                            // Indicates the origin of the cross-origin request.
	TimingAllowOrigin             = "Timing-Allow-Origin"               // Specifies which origins are allowed to access timing information.
	XPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies" // Restricts the loading of certain cross-origin resources.

	// Do Not Track - These headers express the user's preferences for tracking.
	DNT = "DNT" // Indicates the user's tracking preference (Do Not Track).
	Tk  = "Tk"  // Indicates the tracking status of the request.

	// Downloads - This header relates to the downloading of content.
	ContentDisposition = "Content-Disposition" // Specifies the disposition of the content (e.g., inline or attachment).

	// Message body information - Headers that describe the content of the message body.
	ContentEncoding = "Content-Encoding" // Specifies how the content is encoded (e.g., gzip).
	ContentLanguage = "Content-Language" // Specifies the language of the content.
	ContentLength   = "Content-Length"   // Indicates the size of the content in bytes.
	ContentLocation = "Content-Location" // Indicates the location of the resource.
	ContentType     = "Content-Type"     // Specifies the media type of the resource (e.g., text/html).

	// Proxies - Headers that describe information related to proxy servers.
	Forwarded       = "Forwarded"         // Contains information about the client connecting through an intermediary.
	Via             = "Via"               // Shows intermediate protocols and recipients between the client and server.
	XForwardedFor   = "X-Forwarded-For"   // Identifies the originating IP address of the client.
	XForwardedHost  = "X-Forwarded-Host"  // Identifies the original host requested by the client.
	XForwardedProto = "X-Forwarded-Proto" // Indicates the protocol used by the client (e.g., HTTP or HTTPS).

	// Redirects - Headers related to HTTP redirection.
	Location = "Location" // Specifies the URL to redirect the client to.

	// Request context - Headers related to the context of the request.
	From           = "From"            // Contains the email address of the user making the request.
	Host           = "Host"            // Specifies the domain name of the server and the TCP port number.
	Referer        = "Referer"         // Provides the URL of the previous resource that referred the client.
	ReferrerPolicy = "Referrer-Policy" // Governs the referer information sent along with requests.
	UserAgent      = "User-Agent"      // Identifies the user agent (client software) making the request.

	// Response context - Headers related to the context of the response.
	Allow  = "Allow"  // Lists the supported HTTP methods for the resource.
	Server = "Server" // Identifies the server software responding to the request.

	// Range requests - Headers related to partial resource requests and responses.
	AcceptRanges = "Accept-Ranges" // Indicates whether the server supports partial requests.
	ContentRange = "Content-Range" // Specifies the byte range of the partial content being returned.
	IfRange      = "If-Range"      // Ensures partial responses only if the entity hasn't changed.
	Range        = "Range"         // Requests a specific range of bytes from a resource.

	// Security - These headers are used to enforce various security policies and protect web resources.
	ContentSecurityPolicy           = "Content-Security-Policy"             // Defines security policies for the content.
	ContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only" // Used for reporting policy violations without enforcing them.
	CrossOriginResourcePolicy       = "Cross-Origin-Resource-Policy"        // Restricts cross-origin resource access.
	ExpectCT                        = "Expect-CT"                           // Enforces the use of Certificate Transparency.
	FeaturePolicy                   = "Feature-Policy"                      // Controls access to browser features.
	PublicKeyPins                   = "Public-Key-Pins"                     // Enforces a set of public keys for HTTPS connections.
	PublicKeyPinsReportOnly         = "Public-Key-Pins-Report-Only"         // Reports pinning violations without enforcing them.
	StrictTransportSecurity         = "Strict-Transport-Security"           // Enforces secure (HTTPS) connections to the server.
	UpgradeInsecureRequests         = "Upgrade-Insecure-Requests"           // Requests that the server upgrade to a secure connection.
	XContentTypeOptions             = "X-Content-Type-Options"              // Prevents MIME type sniffing.
	XDownloadOptions                = "X-Download-Options"                  // Controls how files are handled during downloads.
	XFrameOptions                   = "X-Frame-Options"                     // Controls whether the browser should allow the page to be framed.
	XPoweredBy                      = "X-Powered-By"                        // Identifies the technology powering the website.
	XXSSProtection                  = "X-XSS-Protection"                    // Enables or disables cross-site scripting (XSS) protection.

	// Server-sent event - These headers are related to the Server-Sent Events (SSE) protocol.
	LastEventID = "Last-Event-ID" // Identifies the last event received from the server in SSE.
	NEL         = "NEL"           // Network Error Logging configuration.
	PingFrom    = "Ping-From"     // The origin initiating the ping request.
	PingTo      = "Ping-To"       // The target URL for the ping request.
	ReportTo    = "Report-To"     // Specifies where to send violation reports.

	// Transfer coding - These headers control transfer encoding behavior.
	TE               = "TE"                // Specifies the transfer encodings the client is willing to accept.
	Trailer          = "Trailer"           // Lists the trailer fields that the client expects to receive after the body.
	TransferEncoding = "Transfer-Encoding" // Specifies the form of encoding used to safely transfer the payload.

	// WebSockets - These headers are related to WebSocket communication.
	SecWebSocketAccept     = "Sec-WebSocket-Accept"     // Server's acceptance of a WebSocket handshake request.
	SecWebSocketExtensions = "Sec-WebSocket-Extensions" // Negotiates WebSocket extensions.
	SecWebSocketKey        = "Sec-WebSocket-Key"        // A key provided by the client to establish the connection.
	SecWebSocketProtocol   = "Sec-WebSocket-Protocol"   // Subprotocols requested by the client.
	SecWebSocketVersion    = "Sec-WebSocket-Version"    // The WebSocket protocol version used by the client.

	// Other - These headers don't fall into the above categories but serve various other purposes.
	AcceptPatch         = "Accept-Patch"           // Indicates which patch document formats the server supports.
	AcceptPushPolicy    = "Accept-Push-Policy"     // Specifies the client's preference for receiving push messages.
	AcceptSignature     = "Accept-Signature"       // Indicates supported signature algorithms.
	AltSvc              = "Alt-Svc"                // Indicates alternative services available.
	Date                = "Date"                   // The date and time when the message was originated.
	Index               = "Index"                  // Specifies the index for specific operations.
	LargeAllocation     = "Large-Allocation"       // Signals the need for a large memory allocation.
	Link                = "Link"                   // Specifies relationships between the current document and other resources.
	PushPolicy          = "Push-Policy"            // Specifies how the server should handle push resources.
	RetryAfter          = "Retry-After"            // Indicates when the client can retry the request after a failure.
	XRatelimitRemaining = "X-Ratelimit-Remaining"  // Shows the number of remaining requests in the current rate limit window.
	ServerTiming        = "Server-Timing"          // Provides metrics on server performance.
	Signature           = "Signature"              // Provides a digital signature for the request.
	SignedHeaders       = "Signed-Headers"         // Lists headers covered by the signature.
	SourceMap           = "SourceMap"              // Points to the source map of a JavaScript resource.
	Upgrade             = "Upgrade"                // Indicates that the client wishes to switch to another protocol.
	XDNSPrefetchControl = "X-DNS-Prefetch-Control" // Controls DNS prefetching.
	XPingback           = "X-Pingback"             // Specifies the URL for pingback.
	XRequestedWith      = "X-Requested-With"       // Identifies requests made via JavaScript libraries.
	XRobotsTag          = "X-Robots-Tag"           // Controls indexing and crawling by web crawlers.
	XUACompatible       = "X-UA-Compatible"        // Specifies the document's compatibility mode for browsers.
)
