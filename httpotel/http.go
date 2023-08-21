package httpotel

import (
	"context"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ClientOption struct {
	Transport   *http.Transport
	Headers     map[string]string
	ContentType string
}

type ClientOptionFunc func(option *ClientOption)

// WithTransport set transport
func (o *ClientOption) WithTransport(transport *http.Transport) ClientOptionFunc {
	return func(option *ClientOption) {
		option.Transport = transport
	}
}

// WithHeaders set headers
func (o *ClientOption) WithHeaders(headers map[string]string) ClientOptionFunc {
	return func(option *ClientOption) {
		option.Headers = headers
	}
}

// WithContentType set content type
func (o *ClientOption) WithContentType(contentType string) ClientOptionFunc {
	return func(option *ClientOption) {
		option.ContentType = contentType
	}
}

func CallApi(ctx context.Context, url string, method string, reqBody io.Reader, timeOut time.Duration, options ...ClientOptionFunc) ([]byte, error) {
	clientOption := &ClientOption{
		Headers: make(map[string]string),
	}

	for _, o := range options {
		o(clientOption)
	}

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   timeOut * time.Second,
	}

	if clientOption.Transport != nil {
		client.Transport = otelhttp.NewTransport(clientOption.Transport)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if clientOption.ContentType != "" {
		req.Header.Set("Content-Type", clientOption.ContentType)
	}

	if len(clientOption.Headers) > 0 {
		for key, data := range clientOption.Headers {
			req.Header.Set(key, data)
		}
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, err

}
