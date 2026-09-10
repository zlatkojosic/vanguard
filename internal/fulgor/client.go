package fulgor

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"nts/vanguard/internal/operators"
)

const DefaultTimeout = 60 * time.Second

type Config struct {
	BaseURL string
	Timeout time.Duration
}

// Client talks to Fulgor's JSON HTTP API.
type Client struct {
	baseURL string
	http    *http.Client

	mu        sync.Mutex
	sessionID string
	operators operators.Overlay
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http:    newHTTPClient(cfg.BaseURL, timeout),
	}
}

func (c *Client) SetSession(id string) {
	c.mu.Lock()
	c.sessionID = id
	c.mu.Unlock()
}

func (c *Client) Session() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

func (c *Client) Operators() operators.Overlay {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.operators
}

func (c *Client) Login(posIdent string) (map[string]string, error) {
	raw, m, err := c.roundTrip(http.MethodPost, "/v2/login", map[string]any{"posIdent": posIdent}, false)
	if err != nil {
		return nil, err
	}
	if sid := m["sessionId"]; sid != "" {
		c.SetSession(sid)
	}
	if m["servisi"] == "" {
		m["servisi"] = m["services"]
	}
	if m["err"] == "60" {
		c.mu.Lock()
		c.operators = ParseOperatorConfig(raw)
		c.mu.Unlock()
	}
	return m, nil
}

func (c *Client) Register(posIdent string) (map[string]string, error) {
	return c.do(http.MethodPost, "/v2/register", map[string]any{"posIdent": posIdent}, true)
}

func (c *Client) Logout() error {
	if c.Session() == "" {
		return nil
	}
	_, err := c.do(http.MethodPost, "/v2/logout", nil, true)
	c.SetSession("")
	return err
}

func (c *Client) Dispatch(fields map[string]string) (map[string]string, error) {
	req, err := resolve(fields)
	if err != nil {
		log.Printf("fulgor skip com=%s: %v", fields["com"], err)
		return map[string]string{"err": "1022"}, nil
	}
	log.Printf("fulgor dispatch com=%s %s %s", fields["com"], req.method, req.path)
	return c.do(req.method, req.path, req.body, true)
}

func (c *Client) do(method, path string, body map[string]any, withSession bool) (map[string]string, error) {
	_, m, err := c.roundTrip(method, path, body, withSession)
	return m, err
}

func (c *Client) roundTrip(method, path string, body map[string]any, withSession bool) ([]byte, map[string]string, error) {
	hasSession := withSession && c.Session() != ""
	start := time.Now()
	log.Printf("fulgor → %s %s%s body=%s", method, c.baseURL+path, sessionTag(hasSession), compactJSON(body))

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		rdr = bytes.NewReader(b)
	}
	httpReq, err := http.NewRequest(method, c.baseURL+path, rdr)
	if err != nil {
		log.Printf("fulgor request error %s %s: %v", method, path, err)
		return nil, nil, err
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if hasSession {
		sid := c.Session()
		httpReq.Header.Set("sessionId", sid)
		httpReq.Header.Set("Authorization", "Bearer "+sid)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		log.Printf("fulgor ← %s %s error after %s: %v", method, path, time.Since(start).Truncate(time.Millisecond), err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("fulgor ← %s %s read error after %s: %v", method, path, time.Since(start).Truncate(time.Millisecond), err)
		return nil, nil, err
	}
	m, err := DecodeBody(raw)
	if err != nil {
		log.Printf("fulgor ← %s %s status=%d dur=%s raw=%s decode=%v", method, path, resp.StatusCode, time.Since(start).Truncate(time.Millisecond), clip(string(raw), 500), err)
		if resp.StatusCode >= 400 {
			return raw, map[string]string{"err": strconv.Itoa(resp.StatusCode)}, nil
		}
		return raw, nil, fmt.Errorf("fulgor: %s", strings.TrimSpace(string(raw)))
	}
	if m["err"] == "" {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			m["err"] = "92"
		case http.StatusForbidden:
			m["err"] = "93"
		case http.StatusTooManyRequests:
			m["err"] = "94"
		}
	}
	log.Printf("fulgor ← %s %s status=%d dur=%s err=%s body=%s", method, path, resp.StatusCode, time.Since(start).Truncate(time.Millisecond), m["err"], clip(string(raw), 500))
	return raw, m, nil
}

func sessionTag(ok bool) string {
	if ok {
		return " session=yes"
	}
	return " session=no"
}

func compactJSON(v map[string]any) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return clip(string(b), 500)
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func BaseURL(scheme, host string, port int) string {
	if scheme == "" {
		scheme = "http"
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

func newHTTPClient(baseURL string, timeout time.Duration) *http.Client {
	c := &http.Client{Timeout: timeout}
	if !skipTLSVerify(baseURL) {
		return c
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // local self-signed Fulgor
	c.Transport = transport
	return c
}

func skipTLSVerify(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme != "https" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}
