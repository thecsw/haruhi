package haruhi

import (
	"net/http"
	"net/url"
)

func mergeParams(dst url.Values, src url.Values) {
	for header, values := range src {
		for _, value := range values {
			dst.Add(header, value)
		}
	}
}

func mergeHeaders(dst http.Header, src http.Header, unset bool) {
	for header, values := range src {
		if unset {
			dst.Del(header)
		}
		for _, value := range values {
			dst.Add(header, value)
		}
	}
}
