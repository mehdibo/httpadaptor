package httpadaptor

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/valyala/fasthttp"
)

func BenchmarkConvertRequest(b *testing.B) {
	var req http.Request

	req.Host = "example.com"
	req.URL = &url.URL{
		Scheme: "http",
		Host:   "example.com",
	}
	req.RequestURI = "/test"
	req.Header = http.Header{}
	req.Header.Set("x", "test")
	req.Header.Set("y", "test")
	req.Method = "GET"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertRequest(&req, &fasthttp.RequestCtx{}, true)
	}
}

func TestConvertRequest(t *testing.T) {
	var req http.Request

	req.Host = "example.com"
	req.URL = &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/test",
	}
	req.RequestURI = "/test"
	req.Header = http.Header{}
	req.Header.Set("X", "test")
	req.Header.Set("Y", "test")
	req.Method = "GET"

	ctx := &fasthttp.RequestCtx{}
	err := ConvertRequest(&req, ctx, true)

	if err != nil {
		t.Fatalf("ConvertRequest returned error: %v", err)
	}

	if string(ctx.Method()) != "GET" {
		t.Errorf("Expected method GET, got %s", ctx.Method())
	}
	if string(ctx.URI().Scheme()) != "http" {
		t.Errorf("Expected scheme http, got %s", ctx.URI().Scheme())
	}
	if string(ctx.Host()) != "example.com" {
		t.Errorf("Expected host example.com, got %s", ctx.Host())
	}
	if string(ctx.RequestURI()) != "/test" {
		t.Errorf("Expected request URI /test, got %s", ctx.RequestURI())
	}
	if string(ctx.Request.Header.Peek("X")) != "test" {
		t.Errorf("Expected header x to be test, got %s", ctx.Request.Header.Peek("x"))
	}
	if string(ctx.Request.Header.Peek("Y")) != "test" {
		t.Errorf("Expected header y to be test, got %s", ctx.Request.Header.Peek("y"))
	}
}
