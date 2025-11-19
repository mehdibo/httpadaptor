package httpadaptor

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

const maxInt = int(^uint(0) >> 1)

var errInvalidRequest = errors.New("httpadaptor: nil request or context")

func ConvertRequest(src *http.Request, dst *fasthttp.RequestCtx, forServer bool) error {
	if src == nil || dst == nil {
		return errInvalidRequest
	}

	req := &dst.Request
	req.Reset()
	req.Header.DisableNormalizing()

	switch {
	case forServer && src.RequestURI != "":
		req.SetRequestURI(src.RequestURI)
	case src.URL != nil:
		populateURIFromRequest(req.URI(), src.URL)
	case src.RequestURI != "":
		req.SetRequestURI(src.RequestURI)
	}

	method := src.Method
	if method == "" {
		method = http.MethodGet
	}
	req.Header.SetMethod(method)

	proto := src.Proto
	if proto == "" {
		proto = "HTTP/1.1"
	}
	req.Header.SetProtocol(proto)

	host := src.Host
	if host == "" && src.URL != nil {
		host = src.URL.Host
	}
	if host != "" {
		req.SetHost(host)
	}

	if src.Close {
		req.Header.SetConnectionClose()
	} else {
		req.Header.ResetConnectionClose()
	}

	hasBody := src.Body != nil && src.Body != http.NoBody
	bodySize := bodySizeFromLength(src.ContentLength)
	if hasBody {
		req.SetBodyStream(src.Body, bodySize)
	} else {
		req.SetBody(nil)
	}

	chunked := isChunked(src.TransferEncoding)
	skipContentLengthHeader := false
	switch {
	case chunked:
		req.Header.SetContentLength(-1)
		skipContentLengthHeader = true
		if len(src.TransferEncoding) > 0 {
			req.Header.Del(fasthttp.HeaderTransferEncoding)
		}
	case bodySize >= 0:
		req.Header.SetContentLength(bodySize)
		skipContentLengthHeader = true
	case hasBody:
		req.Header.SetContentLength(-1)
		skipContentLengthHeader = true
	}

	copyHeaders(&req.Header, src.Header, skipContentLengthHeader)

	dst.SetRemoteAddr(remoteAddrFromRequest(src))

	return nil
}

func copyHeaders(dst *fasthttp.RequestHeader, src http.Header, skipContentLength bool) {
	if len(src) == 0 {
		return
	}
	for name, values := range src {
		if len(values) == 0 {
			continue
		}
		if name == "Host" || (skipContentLength && name == "Content-Length") {
			continue
		}
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

func bodySizeFromLength(n int64) int {
	if n < 0 || n > int64(maxInt) {
		return -1
	}
	return int(n)
}

func isChunked(encodings []string) bool {
	if len(encodings) == 0 {
		return false
	}
	return strings.EqualFold(encodings[len(encodings)-1], "chunked")
}

func populateURIFromRequest(dst *fasthttp.URI, src *url.URL) {
	if dst == nil || src == nil {
		return
	}

	if src.Scheme != "" {
		dst.SetScheme(src.Scheme)
	}

	if src.Host != "" {
		dst.SetHost(src.Host)
	}

	if src.User != nil {
		if username := src.User.Username(); username != "" {
			dst.SetUsername(username)
		} else {
			dst.SetUsername("")
		}
		if password, ok := src.User.Password(); ok {
			dst.SetPassword(password)
		} else {
			dst.SetPassword("")
		}
	} else {
		dst.SetUsername("")
		dst.SetPassword("")
	}

	if src.Fragment != "" {
		dst.SetHash(src.Fragment)
	} else {
		dst.SetHash("")
	}

	if src.Opaque != "" {
		dst.SetPath(src.Opaque)
		dst.SetQueryString("")
		return
	}

	path := src.EscapedPath()
	if path == "" {
		path = "/"
	}
	dst.SetPath(path)
	dst.SetQueryString(src.RawQuery)
}

func remoteAddrFromRequest(r *http.Request) net.Addr {
	if r == nil || r.RemoteAddr == "" {
		return &net.TCPAddr{}
	}

	host, portStr, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip := net.ParseIP(r.RemoteAddr)
		if ip == nil {
			return &net.TCPAddr{}
		}
		return &net.TCPAddr{IP: ip}
	}

	ip := net.ParseIP(host)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 0
	}

	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
}
