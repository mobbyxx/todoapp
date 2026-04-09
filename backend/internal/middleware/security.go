package middleware

import "net/http"

const (
	contentSecurityPolicy     = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; upgrade-insecure-requests;"
	strictTransportSecurity   = "max-age=63072000; includeSubDomains; preload"
	crossOriginEmbedderPolicy = "require-corp"
	crossOriginOpenerPolicy   = "same-origin"
	crossOriginResourcePolicy = "same-origin"
	permissionsPolicy         = "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=(), autoplay=(), display-capture=(), encrypted-media=(), fullscreen=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=()"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("Strict-Transport-Security", strictTransportSecurity)
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("Permissions-Policy", permissionsPolicy)
		w.Header().Set("Cross-Origin-Embedder-Policy", crossOriginEmbedderPolicy)
		w.Header().Set("Cross-Origin-Opener-Policy", crossOriginOpenerPolicy)
		w.Header().Set("Cross-Origin-Resource-Policy", crossOriginResourcePolicy)

		next.ServeHTTP(w, r)
	})
}

func SecurityHeadersWithoutHSTS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("Permissions-Policy", permissionsPolicy)
		w.Header().Set("Cross-Origin-Embedder-Policy", crossOriginEmbedderPolicy)
		w.Header().Set("Cross-Origin-Opener-Policy", crossOriginOpenerPolicy)
		w.Header().Set("Cross-Origin-Resource-Policy", crossOriginResourcePolicy)

		next.ServeHTTP(w, r)
	})
}
