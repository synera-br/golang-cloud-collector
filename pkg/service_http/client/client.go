package http_client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/http2"
)

// IHttpClient interface for http client
type HttpClientInterface interface {
	Get(context.Context, *HttpData) (*http.Response, error)
	Post(context.Context, *HttpData) (*http.Response, error)
}

type HttpData struct {
	BaseURL       string `binding:"required"`
	Body          io.Reader
	Header        map[string]string
	TLS           *tls.Config
	Http2Disabled bool
	client        *http.Client
}

// HttpClient struct of http client
type HttpClient struct {
	Client *http.Client
}

func NewHttpClient(h *HttpClient) (HttpClientInterface, error) {

	return &HttpClient{
		Client: h.Client,
	}, nil
}

func (h *HttpClient) Get(ctx context.Context, data *HttpData) (*http.Response, error) {

	// BaseURL validate
	err := h.parseURI(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, data.BaseURL, nil)
	if err != nil {
		return nil, err
	}

	// Header validate
	err = h.parseHeader(data)
	if err != nil {
		return nil, err
	}

	for k, v := range data.Header {
		req.Header.Add(k, v)
	}

	req = req.WithContext(ctx)

	h.Client = &http.Client{}
	h.setTLS(data)

	res, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *HttpClient) Post(ctx context.Context, data *HttpData) (*http.Response, error) {

	// BaseURL validate
	err := h.parseURI(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, data.BaseURL, data.Body)
	if err != nil {
		return nil, err
	}

	// Header validate
	err = h.parseHeader(data)
	if err != nil {
		return nil, err
	}

	for k, v := range data.Header {
		req.Header.Add(k, v)
	}

	req = req.WithContext(ctx)
	// proxyUrl, _ := url.Parse("http://proxy.sig.umbrella.com:443")

	h.Client = &http.Client{
		// Transport: &http.Transport{
		// 	Proxy: http.ProxyURL(proxyUrl),
		// },
	}

	h.setTLS(data)

	res, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (h *HttpClient) parseHeader(data *HttpData) error {

	for k, v := range data.Header {
		if v == "" {
			return fmt.Errorf("the %s field cannot be empty", k)
		}
	}
	return nil
}

func (h *HttpClient) parseURI(data *HttpData) error {

	_, err := url.ParseRequestURI(data.BaseURL)
	if err != nil {
		return err
	}
	return nil
}

func (h *HttpClient) setTLS(data *HttpData) {
	// TLS validate
	if data.TLS != nil {
		if !data.Http2Disabled {
			h.Client.Transport = &http2.Transport{
				TLSClientConfig: data.TLS,
			}
		}

		if data.Http2Disabled {
			h.Client.Transport = &http.Transport{
				TLSClientConfig: data.TLS,
			}
		}
	}
}
