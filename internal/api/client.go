package api

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxBodySize = 10 * 1024 * 1024

type RequestResult struct {
	URL         string
	StatusCode  int
	StatusText  string
	Duration    time.Duration
	Headers     http.Header
	Body        []byte
	BodySize    int64
	ContentType string
	Proto       string
	Error       error
	Timestamp   time.Time
}

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				DisableCompression: false,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
			},
			Timeout: timeout,
		},
	}
}

func (c *Client) Get(rawURL string, customHeaders map[string]string) RequestResult {
	res := RequestResult{
		URL:       rawURL,
		Timestamp: time.Now(),
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
		res.URL = rawURL
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		res.Error = fmt.Errorf("invalid request: %w", err)
		return res
	}

	req.Header.Set("User-Agent", "LazyAPI/1.0 (Terminal; +https://github.com/jan/lazyapi)")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	for k, v := range customHeaders {
		kTrim := strings.TrimSpace(k)
		if kTrim != "" {
			req.Header.Set(kTrim, strings.TrimSpace(v))
		}
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	res.Duration = time.Since(start)

	if err != nil {
		res.Error = err
		return res
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	res.StatusText = resp.Status
	res.Headers = resp.Header
	res.Proto = resp.Proto
	res.ContentType = resp.Header.Get("Content-Type")

	limitedReader := io.LimitReader(resp.Body, maxBodySize)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		res.Error = fmt.Errorf("reading body failed: %w", err)
		return res
	}

	res.Body = bodyBytes
	res.BodySize = int64(len(bodyBytes))
	return res
}
