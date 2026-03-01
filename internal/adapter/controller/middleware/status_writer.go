package middleware

import "net/http"

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newStatusWriter(w http.ResponseWriter) *statusWriter {
	return &statusWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (sw *statusWriter) Status() int {
	return sw.status
}

func (sw *statusWriter) WriteHeader(code int) {
	if sw.wroteHeader {
		return
	}

	sw.status = code
	sw.wroteHeader = true

	sw.ResponseWriter.WriteHeader(code)
}
