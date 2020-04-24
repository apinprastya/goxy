package server

import (
	"net/http"
)

var defaultHTML = []byte(`
<html>
<head><title>GOXY Reverse Proxy</title></head>
<body>
    <center>
        <h2>GOXY Reverse Proxy is HERE!!!</h2><br>
        <a href="https://github.com/apinprastya/goxy">https://github.com/apinprastya/goxy</a>
    </center>
</body>
</html>`)

func defaultHTTPResponse(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(defaultHTML)
}
