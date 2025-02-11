package header

// Header represents an HTTP header field as defined by IANA.
// It is implemented as a string type to facilitate easy usage and type safety.
// Reference: https://www.iana.org/assignments/http-fields/http-fields.xhtml
//
// The `Header` type provides a strongly typed way to work with standard HTTP headers,
// preventing typos and improving code maintainability.
type Header string

// String returns the string representation of the Header type.
//
// Returns:
//   - header (string): The HTTP header as a string value.
//
// Example:
//
//	h := Authorization
//	fmt.Println(h.String()) // Output: "Authorization"
func (h Header) String() (header string) {
	return string(h)
}

// Interface defines a common interface for Header types.
// Any type that provides a String() method returning a string may be used
// interchangeably where a Header-like type is required.
//
// This is particularly useful if you wish to define additional types or wrappers
// that adhere to the same contract as Header.
type Interface interface {
	String() (header string)
}

// Authentication & Authorization
// These constants represent header fields related to authentication and authorization.
// They are used to pass credentials and to signal the required authentication schemes.
const (
	Authorization      Header = "Authorization"       // Used to pass authentication credentials (e.g., bearer tokens).
	ProxyAuthenticate  Header = "Proxy-Authenticate"  // Sent in a response indicating that the client must authenticate with the proxy.
	ProxyAuthorization Header = "Proxy-Authorization" // Used in requests to provide authentication credentials to a proxy.
	WWWAuthenticate    Header = "WWW-Authenticate"    // Sent by the server to indicate the authentication method(s) that should be used.
)

// Caching
// These headers control caching behavior between clients and servers. They specify
// cache lifetimes, validation conditions, and other cache directives.
const (
	Age           Header = "Age"             // Indicates the age (in seconds) of a cached response.
	CacheControl  Header = "Cache-Control"   // Contains directives for caching mechanisms.
	ClearSiteData Header = "Clear-Site-Data" // Instructs the client to clear cached data.
	Expires       Header = "Expires"         // Specifies the date/time after which the response is considered stale.
	Pragma        Header = "Pragma"          // Provides backwards-compatible caching directives.
	Warning       Header = "Warning"         // Carries additional information about the message's status or transformation.
)

// Client Hints
// These headers provide the server with hints about the client's environment
// (e.g., device characteristics, user preferences), allowing for an adaptive response.
const (
	AcceptCH         Header = "Accept-CH"          // Indicates which client hints the server supports.
	AcceptCHLifetime Header = "Accept-CH-Lifetime" // Specifies the duration for which the client should retain client hint preferences.
	ContentDPR       Header = "Content-DPR"        // Provides the device pixel ratio for rendering purposes.
	DPR              Header = "DPR"                // A shorthand for the device pixel ratio.
	EarlyData        Header = "Early-Data"         // Signals that the request is using early data (such as TLS 1.3 0-RTT).
	SaveData         Header = "Save-Data"          // Indicates that the client prefers to minimize data usage.
	ViewportWidth    Header = "Viewport-Width"     // Provides the width of the viewport.
	Width            Header = "Width"              // Specifies the display width of the client’s device.
)

// Conditionals
// These headers are used to make HTTP requests conditional, allowing operations
// such as fetching or updating a resource only if certain conditions are met.
const (
	ETag              Header = "ETag"                // Provides a unique identifier for a specific version of a resource.
	IfMatch           Header = "If-Match"            // The request is performed only if the provided ETag matches.
	IfModifiedSince   Header = "If-Modified-Since"   // The request is performed only if the resource has been modified after the specified date.
	IfNoneMatch       Header = "If-None-Match"       // The request is performed only if the provided ETag does not match.
	IfUnmodifiedSince Header = "If-Unmodified-Since" // The request is performed only if the resource has not been modified since the given date.
	LastModified      Header = "Last-Modified"       // Indicates the date and time the resource was last modified.
	Vary              Header = "Vary"                // Specifies which request headers should be considered when caching responses.
)

// Connection management
// These headers help manage the behavior of the network connection between the client and server.
const (
	Connection      Header = "Connection"       // Controls options for the current connection (e.g., keep-alive).
	KeepAlive       Header = "Keep-Alive"       // Provides parameters for maintaining a persistent connection.
	ProxyConnection Header = "Proxy-Connection" // Specifies connection options when communicating via a proxy.
)

// Content negotiation
// These headers enable the client to specify the desired media types, encodings, and languages
// for the response.
const (
	Accept         Header = "Accept"          // Indicates the media types that are acceptable for the response.
	AcceptCharset  Header = "Accept-Charset"  // Specifies the character sets that are acceptable.
	AcceptEncoding Header = "Accept-Encoding" // Lists the acceptable content encodings (e.g., gzip, deflate).
	AcceptLanguage Header = "Accept-Language" // States the preferred languages for the response.
)

// Controls
// General-purpose headers used to control aspects of the request or response,
// such as cookies and expectations about processing.
const (
	Cookie      Header = "Cookie"       // Contains stored cookies sent from the client to the server.
	Expect      Header = "Expect"       // Indicates that particular conditions must be met by the server.
	MaxForwards Header = "Max-Forwards" // Limits the number of times a request can be forwarded.
	SetCookie   Header = "Set-Cookie"   // Sent by the server to instruct the client to store cookies.
)

// CORS (Cross-Origin Resource Sharing)
// These headers manage cross-origin requests and responses, specifying which origins,
// methods, and headers are permitted.
const (
	AccessControlAllowCredentials Header = "Access-Control-Allow-Credentials"  // Indicates whether credentials are allowed in cross-origin requests.
	AccessControlAllowHeaders     Header = "Access-Control-Allow-Headers"      // Lists the headers allowed in cross-origin requests.
	AccessControlAllowMethods     Header = "Access-Control-Allow-Methods"      // Specifies the HTTP methods permitted for cross-origin requests.
	AccessControlAllowOrigin      Header = "Access-Control-Allow-Origin"       // Defines the origin(s) allowed to access the resource.
	AccessControlExposeHeaders    Header = "Access-Control-Expose-Headers"     // Identifies which headers can be exposed to the client.
	AccessControlMaxAge           Header = "Access-Control-Max-Age"            // States how long the results of a preflight request can be cached.
	AccessControlRequestHeaders   Header = "Access-Control-Request-Headers"    // Specifies which headers will be used in the actual request.
	AccessControlRequestMethod    Header = "Access-Control-Request-Method"     // Specifies the method to be used in the actual request.
	Origin                        Header = "Origin"                            // Indicates the origin (scheme, host, port) of the request.
	TimingAllowOrigin             Header = "Timing-Allow-Origin"               // Specifies which origins can access timing information.
	XPermittedCrossDomainPolicies Header = "X-Permitted-Cross-Domain-Policies" // Restricts loading of cross-domain resources.
)

// Do Not Track
// These headers communicate the user’s preference regarding tracking.
const (
	DNT Header = "DNT" // Indicates the user’s tracking preference (Do Not Track).
	Tk  Header = "Tk"  // Conveys the tracking status of the request.
)

// Downloads
// These headers control how content should be handled by the client, especially regarding downloads.
const (
	ContentDisposition Header = "Content-Disposition" // Specifies whether the content should be displayed inline or treated as an attachment.
)

// Message body information
// These headers describe the properties of the message body, such as its encoding,
// language, size, and media type.
const (
	ContentEncoding Header = "Content-Encoding" // Specifies the encoding transformation applied to the message body (e.g., gzip).
	ContentLanguage Header = "Content-Language" // Indicates the natural language(s) of the message body.
	ContentLength   Header = "Content-Length"   // Defines the size of the message body in bytes.
	ContentLocation Header = "Content-Location" // Provides a URI for an alternate location of the returned data.
	ContentType     Header = "Content-Type"     // Indicates the media type of the resource (e.g., text/html).
)

// Proxies
// These headers are relevant when a request is passed through one or more proxy servers,
// conveying information about the client’s original request.
const (
	Forwarded       Header = "Forwarded"         // Contains information about the client and any intermediate proxies.
	Via             Header = "Via"               // Lists intermediate protocols and recipients between the client and server.
	XForwardedFor   Header = "X-Forwarded-For"   // Identifies the originating IP address of the client connecting through a proxy.
	XForwardedHost  Header = "X-Forwarded-Host"  // Specifies the original host requested by the client.
	XForwardedProto Header = "X-Forwarded-Proto" // Indicates the protocol (HTTP or HTTPS) used by the client.
)

// Redirects
// These headers are used to manage HTTP redirection, guiding the client to the appropriate resource.
const (
	Location Header = "Location" // Provides the URL to which the client should be redirected.
)

// Request context
// These headers supply additional context about the request, including information about the client
// and its origin.
const (
	From           Header = "From"            // Contains the email address of the user making the request.
	Host           Header = "Host"            // Specifies the host and port number to which the request is directed.
	Referer        Header = "Referer"         // Indicates the URL of the web page from which the request originated.
	ReferrerPolicy Header = "Referrer-Policy" // Governs the amount of referrer information sent with requests.
	UserAgent      Header = "User-Agent"      // Identifies the client software (e.g., browser) making the request.
)

// Response context
// These headers provide information about the server and the capabilities of the requested resource.
const (
	Allow  Header = "Allow"  // Lists the HTTP methods that are supported for the requested resource.
	Server Header = "Server" // Contains information about the software used by the origin server.
)

// Range requests
// These headers facilitate partial retrieval of resources, allowing the client to request only a portion
// of a resource rather than the entire payload.
const (
	AcceptRanges Header = "Accept-Ranges" // Indicates whether the server supports partial requests.
	ContentRange Header = "Content-Range" // Specifies the byte range of the content being delivered.
	IfRange      Header = "If-Range"      // Makes the request conditional: the server returns the range only if the resource is unchanged.
	Range        Header = "Range"         // Requests a specific portion of the resource by specifying a byte range.
)

// Security
// These headers enforce various security policies and help protect both clients and servers from attacks.
const (
	ContentSecurityPolicy           Header = "Content-Security-Policy"             // Defines content security policies to mitigate cross-site scripting (XSS) and other attacks.
	ContentSecurityPolicyReportOnly Header = "Content-Security-Policy-Report-Only" // Specifies a policy for reporting violations without enforcing them.
	CrossOriginResourcePolicy       Header = "Cross-Origin-Resource-Policy"        // Restricts the ways resources can be loaded from other origins.
	ExpectCT                        Header = "Expect-CT"                           // Enforces Certificate Transparency for secure connections.
	FeaturePolicy                   Header = "Feature-Policy"                      // Controls the use of browser features (deprecated in favor of Permissions Policy).
	PublicKeyPins                   Header = "Public-Key-Pins"                     // Specifies acceptable public keys for HTTPS connections (largely deprecated).
	PublicKeyPinsReportOnly         Header = "Public-Key-Pins-Report-Only"         // Similar to PublicKeyPins but only reports violations.
	StrictTransportSecurity         Header = "Strict-Transport-Security"           // Enforces secure (HTTPS) connections to the server.
	UpgradeInsecureRequests         Header = "Upgrade-Insecure-Requests"           // Instructs the browser to upgrade all HTTP requests to HTTPS.
	XContentTypeOptions             Header = "X-Content-Type-Options"              // Prevents MIME-sniffing by the browser.
	XDownloadOptions                Header = "X-Download-Options"                  // Controls file download behavior in certain browsers.
	XFrameOptions                   Header = "X-Frame-Options"                     // Protects against clickjacking by controlling whether the page can be framed.
	XPoweredBy                      Header = "X-Powered-By"                        // Reveals information about the underlying technology used by the server.
	XXSSProtection                  Header = "X-XSS-Protection"                    // Enables built-in cross-site scripting (XSS) filtering in some browsers.
)

// Server-Sent Events (SSE)
// These headers are associated with Server-Sent Events, a mechanism that enables the server to push
// real-time updates to the client.
const (
	LastEventID Header = "Last-Event-ID" // Provides the ID of the last event received, allowing the client to resume correctly.
	NEL         Header = "NEL"           // Configures Network Error Logging.
	PingFrom    Header = "Ping-From"     // Specifies the origin initiating a ping in an SSE context.
	PingTo      Header = "Ping-To"       // Specifies the target URL for an SSE ping.
	ReportTo    Header = "Report-To"     // Instructs where to send reports of network errors or policy violations.
)

// Transfer Coding
// These headers control the transfer encoding mechanisms used for the payload,
// such as chunked or compressed data.
const (
	TE               Header = "TE"                // Specifies acceptable transfer encodings (e.g., chunked).
	Trailer          Header = "Trailer"           // Lists the trailer fields to be sent after the message body.
	TransferEncoding Header = "Transfer-Encoding" // Indicates the encoding (e.g., chunked) used to safely transfer the payload.
)

// WebSockets
// These headers are used during the WebSocket handshake and connection upgrade process,
// facilitating the establishment of a full-duplex communication channel.
const (
	SecWebSocketAccept     Header = "Sec-WebSocket-Accept"     // The server’s response value to confirm the WebSocket handshake.
	SecWebSocketExtensions Header = "Sec-WebSocket-Extensions" // Specifies any WebSocket protocol extensions in use.
	SecWebSocketKey        Header = "Sec-WebSocket-Key"        // A unique key sent by the client to initiate the WebSocket handshake.
	SecWebSocketProtocol   Header = "Sec-WebSocket-Protocol"   // Indicates the subprotocol(s) the client wishes to use.
	SecWebSocketVersion    Header = "Sec-WebSocket-Version"    // Specifies the WebSocket protocol version used by the client.
)

// Miscellaneous
// These headers serve various other purposes and do not belong to a single category.
const (
	AcceptPatch         Header = "Accept-Patch"           // Indicates which patch document formats are supported.
	AcceptPushPolicy    Header = "Accept-Push-Policy"     // Specifies the client’s preference for push notifications.
	AcceptSignature     Header = "Accept-Signature"       // Lists supported algorithms for request signature verification.
	AltSvc              Header = "Alt-Svc"                // Advertises alternative services (e.g., HTTP/2, QUIC).
	Date                Header = "Date"                   // Provides the date and time at which the message was originated.
	Index               Header = "Index"                  // Can be used for operations that require an index reference.
	LargeAllocation     Header = "Large-Allocation"       // Signals that a large memory allocation may be necessary.
	Link                Header = "Link"                   // Specifies relationships between the current document and other resources.
	PushPolicy          Header = "Push-Policy"            // Instructs how server push should be handled.
	RetryAfter          Header = "Retry-After"            // Indicates the delay before the client should retry the request.
	XRatelimitRemaining Header = "X-Ratelimit-Remaining"  // Shows the number of remaining requests in the current rate limiting window.
	ServerTiming        Header = "Server-Timing"          // Provides performance metrics for the server's processing of the request.
	Signature           Header = "Signature"              // Contains a digital signature to verify the authenticity of the request.
	SignedHeaders       Header = "Signed-Headers"         // Lists which headers have been signed.
	SourceMap           Header = "SourceMap"              // Provides a URL to the source map for debugging JavaScript.
	Upgrade             Header = "Upgrade"                // Requests that the server switch protocols (e.g., to WebSocket).
	XDNSPrefetchControl Header = "X-DNS-Prefetch-Control" // Controls DNS prefetching behavior.
	XPingback           Header = "X-Pingback"             // Specifies a URL for pingback notifications.
	XRequestedWith      Header = "X-Requested-With"       // Identifies the client making the request, commonly used with AJAX.
	XRobotsTag          Header = "X-Robots-Tag"           // Provides directives to web crawlers regarding indexing and following links.
	XUACompatible       Header = "X-UA-Compatible"        // Specifies compatibility modes for browsers (primarily for Internet Explorer).
)

// This compile-time assertion ensures that the Header type correctly implements the Interface interface.
// If it does not, the assignment will cause a compile-time error.
var _ Interface = (*Header)(nil)
