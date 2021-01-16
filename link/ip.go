package link

import (
	"net"
	"net/http"
)

func RemoteIP(r *http.Request) (remote string) {
	remote = r.Header.Get("X-Real-Ip")
	if remote == "" {
		remote = r.Header.Get("X-Forwarded-For")
	}
	if remote == "" {
		remote = r.RemoteAddr
	}
	remote, _, _ = net.SplitHostPort(remote)
	return
}
