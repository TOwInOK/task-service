package http

import "net/http"

// chiURLParam extracts a URL parameter using chi's path params.
// Falls back to query parameter if chi context is not available.
func chiURLParam(r *http.Request, key string) string {
	if chiCtx := chiRouteContext(r); chiCtx != nil {
		if v := chiCtx.URLParam(key); v != "" {
			return v
		}
	}
	return r.URL.Query().Get(key)
}
