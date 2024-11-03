package packet

import (
	"bufio"
	"bytes"
)

var (
	httpMethods = []string{
		"HEAD",
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
		"CONNECT",
		"TRACE",
		"PATCH",
	}
)

// IsHTTP returns true if the data starts with an HTTP method.
func IsHTTP(r *bufio.Reader) bool {
	peek, err := r.Peek(4)
	if err != nil {
		return false
	}
	for _, method := range httpMethods {
		if bytes.HasPrefix(peek, []byte(method)) {
			return true
		}
	}
	return false
}
