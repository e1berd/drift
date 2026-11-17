package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// PublicPrefix is the path under which the Kratos public API is exposed on the application origin,
// so that Kratos cookies, CSRF tokens and WebAuthn origins all belong to one site.
const PublicPrefix = "/.ory/kratos/public"

// ErrFlowExpired is returned when a flow no longer exists, has expired or belongs to another browser.
var ErrFlowExpired = errors.New("kratos: flow expired or invalid")

// APIError is a non-2xx response from Kratos.
type APIError struct {
	Status  int
	ID      string `json:"id"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kratos: %d %s: %s %s", e.Status, e.ID, e.Message, e.Reason)
}

// Client talks to the Kratos public and admin APIs.
type Client struct {
	public *url.URL
	admin  *url.URL
	http   *http.Client
}

// New returns a client for the given Kratos public and admin base URLs.
func New(public, admin *url.URL) *Client {
	return &Client{
		public: public,
		admin:  admin,
		http: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// HasIdentities reports whether at least one identity exists.
func (c *Client) HasIdentities(ctx context.Context) (bool, error) {
	var ids []Identity
	if err := c.do(ctx, c.admin, http.MethodGet, "/admin/identities?page_size=1", "", nil, &ids); err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

// CreateIdentityParams describes an identity created through the admin API.
type CreateIdentityParams struct {
	Traits         Traits
	Password       string
	MetadataPublic map[string]any
}

// CreateIdentity creates an active identity with a password credential.
func (c *Client) CreateIdentity(ctx context.Context, p CreateIdentityParams) (*Identity, error) {
	body := map[string]any{
		"schema_id": "default",
		"state":     "active",
		"traits":    p.Traits,
		"credentials": map[string]any{
			"password": map[string]any{"config": map[string]any{"password": p.Password}},
		},
		"metadata_public": p.MetadataPublic,
	}
	var id Identity
	if err := c.do(ctx, c.admin, http.MethodPost, "/admin/identities", "", body, &id); err != nil {
		return nil, err
	}
	return &id, nil
}

// Session returns the session for the given Cookie header, or nil when the caller is not signed in.
func (c *Client) Session(ctx context.Context, cookie string) (*Session, error) {
	if cookie == "" {
		return nil, nil
	}
	var s Session
	err := c.do(ctx, c.public, http.MethodGet, "/sessions/whoami", cookie, nil, &s)
	var apiErr *APIError
	if errors.As(err, &apiErr) && (apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !s.Active {
		return nil, nil
	}
	return &s, nil
}

// LoginFlow fetches a browser login flow on behalf of the browser that owns the Cookie header.
func (c *Client) LoginFlow(ctx context.Context, id, cookie string) (*Flow, error) {
	var f Flow
	err := c.do(ctx, c.public, http.MethodGet, "/self-service/login/flows?id="+url.QueryEscape(id), cookie, nil, &f)
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusNotFound, http.StatusGone, http.StatusForbidden, http.StatusBadRequest:
			return nil, ErrFlowExpired
		}
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// SelfServiceError fetches a self-service error by id.
func (c *Client) SelfServiceError(ctx context.Context, id string) (*FlowError, error) {
	var e FlowError
	if err := c.do(ctx, c.public, http.MethodGet, "/self-service/errors?id="+url.QueryEscape(id), "", nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// LogoutURL creates a browser logout flow and returns the URL that ends the session.
func (c *Client) LogoutURL(ctx context.Context, cookie string) (string, error) {
	var out struct {
		LogoutURL string `json:"logout_url"`
	}
	if err := c.do(ctx, c.public, http.MethodGet, "/self-service/logout/browser", cookie, nil, &out); err != nil {
		return "", err
	}
	return out.LogoutURL, nil
}

// PublicProxy forwards requests under PublicPrefix to the Kratos public API.
func (c *Client) PublicProxy() http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(c.public)
			r.SetXForwarded()
			r.Out.URL.Path = strings.TrimPrefix(r.In.URL.Path, PublicPrefix)
			r.Out.URL.RawPath = strings.TrimPrefix(r.In.URL.RawPath, PublicPrefix)
			if r.Out.URL.Path == "" {
				r.Out.URL.Path = "/"
			}
		},
	}
}

func (c *Client) do(ctx context.Context, base *url.URL, method, path, cookie string, in, out any) error {
	ref, err := url.Parse(path)
	if err != nil {
		return err
	}
	target := base.JoinPath(ref.Path)
	target.RawQuery = ref.RawQuery

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("kratos: %s %s: %w", method, ref.Path, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		apiErr := &APIError{Status: res.StatusCode}
		var wrapped struct {
			Error *APIError `json:"error"`
		}
		if json.Unmarshal(data, &wrapped) == nil && wrapped.Error != nil {
			apiErr = wrapped.Error
			apiErr.Status = res.StatusCode
		}
		return apiErr
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}
