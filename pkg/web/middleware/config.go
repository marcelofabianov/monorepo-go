package middleware

type SecurityHeadersConfig struct {
	XContentTypeOptions      string
	XFrameOptions            string
	ContentSecurityPolicy    string
	ReferrerPolicy           string
	StrictTransportSecurity  string
	CacheControl             string
	PermissionsPolicy        string
	XDNSPrefetchControl      string
	XDownloadOptions         string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}
