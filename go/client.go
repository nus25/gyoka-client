package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nus25/gyokaclient/go/types"
)

// HTTPクライアントの設定を定数として定義
const (
	defaultTimeout      = 30 * time.Second
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
	idleConnTimeout     = 90 * time.Second
	maxRetries          = 3
	retryWaitTime       = 1 * time.Second
)

type ClientOptions struct {
	Timeout       time.Duration
	MaxIdleConns  int
	MaxRetries    int
	RetryWaitTime time.Duration
	Auth          *AuthConfig
}

func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Timeout:       defaultTimeout,
		MaxIdleConns:  maxIdleConns,
		MaxRetries:    maxRetries,
		RetryWaitTime: retryWaitTime,
		Auth: &AuthConfig{
			Type:        NoAuth,
			Credentials: make(map[string]string),
		},
	}
}

type BadRequestResponse struct {
	Message string `json:"message"`
}

type ListPostResponse struct {
	Feed  string       `json:"feed"`
	Count int          `json:"count"`
	Posts []types.Post `json:"posts"`
}

type CreatePostRequest struct {
	Posts []types.Post `json:"posts"`
}

type CreatePostResponse struct {
	InsertedPosts []types.Post `json:"insertedPosts"`
	FailedPosts   []types.Post `json:"failedPosts"`
	Message       string       `json:"message"`
}

type DeletePostRequest struct {
	Posts []types.Post `json:"posts"`
}

type DeletePostResponse struct {
	DeletedPosts []types.Post `json:"deletedPosts"`
	FailedPosts  []types.Post `json:"failedPosts"`
	Message      string       `json:"message"`
}
type TrimResponse struct {
	Message      string `json:"message"`
	DeletedCount int    `json:"deletedCount"`
}

type AuthType int

const (
	NoAuth AuthType = iota
	CloudflareAccess
	BearerToken
	BasicAuth
)

func authTypeToString(t AuthType) string {
	switch t {
	case NoAuth:
		return "NoAuth"
	case CloudflareAccess:
		return "CloudflareAccess"
	case BearerToken:
		return "BearerToken"
	case BasicAuth:
		return "BasicAuth"
	default:
		return "Unknown"
	}
}

type AuthConfig struct {
	Type        AuthType
	Credentials map[string]string
}

type Client struct {
	BaseURL    string
	logger     *slog.Logger
	httpClient *http.Client
	config     *ClientOptions
}

func NewClient(baseURL string, logger *slog.Logger, opts ...ClientOption) (*Client, error) {
	opt := DefaultClientOptions()
	for _, o := range opts {
		o(opt)
	}
	logger = logger.With("component", "gyoka-editor-client")
	logger.Info("creating new client",
		"baseURL", baseURL,
		"authType", authTypeToString(opt.Auth.Type))

	transport := &http.Transport{
		MaxIdleConns:        opt.MaxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}

	client := &Client{
		BaseURL: baseURL,
		logger:  logger,
		httpClient: &http.Client{
			Timeout:   opt.Timeout,
			Transport: transport,
		},
		config: opt,
	}

	return client, nil
}

type ClientOption func(*ClientOptions)

func WithToken(token string) ClientOption {
	return func(opts *ClientOptions) {
		if opts.Auth.Type != NoAuth {
			return
		}
		opts.Auth.Type = BearerToken
		opts.Auth.Credentials["token"] = token
	}
}

func WithCloudflareAccess(clientID, clientSecret string) ClientOption {
	return func(opts *ClientOptions) {
		if opts.Auth.Type != NoAuth {
			return
		}
		opts.Auth.Type = CloudflareAccess
		opts.Auth.Credentials["clientId"] = clientID
		opts.Auth.Credentials["clientSecret"] = clientSecret
	}
}

func WithBasicAuth(username, password string) ClientOption {
	return func(opts *ClientOptions) {
		if opts.Auth.Type != NoAuth {
			return
		}
		opts.Auth.Type = BasicAuth
		opts.Auth.Credentials["username"] = username
		opts.Auth.Credentials["password"] = password
	}
}

func (c *Client) GetAuthType() AuthType {
	return c.config.Auth.Type
}

func (c *Client) applyAuth(req *http.Request) {
	switch c.config.Auth.Type {
	case BearerToken:
		if token := c.config.Auth.Credentials["token"]; token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case CloudflareAccess:
		req.Header.Set("CF-Access-Client-Id", c.config.Auth.Credentials["clientId"])
		req.Header.Set("CF-Access-Client-Secret", c.config.Auth.Credentials["clientSecret"])
	case BasicAuth:
		req.SetBasicAuth(
			c.config.Auth.Credentials["username"],
			c.config.Auth.Credentials["password"],
		)
	}
}

// リトライ処理を含むHTTPリクエスト実行関数
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("retrying request",
				"attempt", attempt+1,
				"url", req.URL.String(),
				"error", lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryWaitTime * time.Duration(attempt)):
			}
		}

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %s", resp.Status)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// レスポンスを処理するための新しいメソッド
func (c *Client) RequestWithResponse(ctx context.Context, method string, path string, body, response interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		c.logger.Error("request failed",
			"method", method,
			"path", path,
			"error", err)
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResponse struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(bodyBytes, &errResponse); err != nil {
			return fmt.Errorf("unexpected status %s: %s", resp.Status, string(bodyBytes))
		}
		return fmt.Errorf("api error: %s", errResponse.Message)
	}

	if err := json.Unmarshal(bodyBytes, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug("request successful",
		"method", method,
		"path", path,
		"status", resp.Status)

	return nil
}

// フィードにポストを追加
func (c *Client) Add(ctx context.Context, posts []types.Post) (*CreatePostResponse, error) {
	if len(posts) == 0 {
		return &CreatePostResponse{}, nil
	}
	if len(posts) > 40 {
		return nil, fmt.Errorf("too many posts: %d (max 40)", len(posts))
	}

	var response CreatePostResponse
	req := CreatePostRequest{Posts: posts}
	if err := c.RequestWithResponse(ctx, http.MethodPost, "/feed/add", req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// フィードからポストを削除
func (c *Client) Delete(ctx context.Context, posts []types.Post) (*DeletePostResponse, error) {
	if len(posts) == 0 {
		return &DeletePostResponse{}, nil
	}
	if len(posts) > 40 {
		return nil, fmt.Errorf("too many posts: %d (max 40)", len(posts))
	}

	var response DeletePostResponse
	req := DeletePostRequest{Posts: posts}
	if err := c.RequestWithResponse(ctx, http.MethodPost, "/feed/delete", req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// フィードを指定した数に切り詰める
func (c *Client) TrimWithCount(ctx context.Context, feed types.FeedUri, count int) (*TrimResponse, error) {
	if count < 0 {
		return nil, fmt.Errorf("invalid count: %d", count)
	}

	if err := feed.Validate(); err != nil {
		return nil, fmt.Errorf("invalid feed uri: %w", err)
	}

	path := fmt.Sprintf("/feed/trim?feed=%s&within-count=%d", feed, count)
	var response TrimResponse
	if err := c.RequestWithResponse(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// フィードからポスト一覧を取得
func (c *Client) ListPost(ctx context.Context, feed types.FeedUri, limit int) (*ListPostResponse, error) {
	if err := feed.Validate(); err != nil {
		return nil, fmt.Errorf("invalid feed uri: %w", err)
	}

	query := fmt.Sprintf("/feed/list?feed=%s", feed)
	if limit > 0 {
		query = fmt.Sprintf("%s&limit=%d", query, limit)
	}

	var response ListPostResponse
	if err := c.RequestWithResponse(ctx, http.MethodGet, query, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// サーバーの生存確認
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.doRequest(ctx, &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   strings.TrimPrefix(strings.TrimPrefix(c.BaseURL, "http://"), "https://"),
			Path:   "/",
		},
		Header: make(http.Header),
	})
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// クライアントの終了処理
func (c *Client) Close() {
	if c.httpClient != nil {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
