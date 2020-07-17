package server

import "net/http"

type httpsHandler struct{}

func newHTTPSHandler() *httpsHandler {
	h := &httpsHandler{}
	return h
}

func (h *httpsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if c, ok := serverConfigMaps[req.Host]; ok {
		c.ServeHTTP(w, req)
	} else {
		defaultHTTPResponse(w, req)
	}
}
