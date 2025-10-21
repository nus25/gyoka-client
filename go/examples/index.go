package main

// Example code for using this client.
// go run examples/index.go
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	client "github.com/nus25/gyoka-client/go"
)

const (
	defaultTimeout      = 30 * time.Second
	maxIdleConns        = 10
	maxIdleConnsPerHost = 10
	idleConnTimeout     = 90 * time.Second
	maxRetries          = 3
	retryWaitTime       = 1 * time.Second
)

func main() {
	// init client
	baseTransport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}
	// custom headers for auth(if configured)
	ch := make(map[string]string)
	// gyoka API key
	ch["X-API-Key"] = "your-api-key"
	// Cloudflare zerotrust service token
	ch["CF-Access-Client-Id"] = "your-client-id"
	ch["CF-Access-Client-Secret"] = "your-client-secret"
	hc := &http.Client{
		Transport: &customHeaderTransport{
			customHeaders: ch,
			transport:     baseTransport,
		},
		Timeout: defaultTimeout,
	}
	cl, err := client.NewClientWithResponses("http://localhost:8787", client.WithHTTPClient(hc))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// compose AddPost request
	indexedAt := time.Now().UTC()
	repostUri := "at://did:plc:1234abcd/app.bsky.feed.repost/record123"
	body := client.PostAddPostJSONRequestBody{
		Feed: "at://did:plc:1234abcd/app.bsky.feed.generator/record123",
		Post: client.AddPostPostParam{
			Uri:       "at://did:plc:1234abcd/app.bsky.feed.post/record123",
			Cid:       "sampleiaajksfnn3if2crogjkz5c4bmb2lh2ufspcdf6hfc7mtg6e2bysva",
			Languages: &[]string{"en"},
			IndexedAt: &indexedAt,
			Reason: &client.AddPostReasonParam{
				Type:   client.AddPostReasonParamTypeAppBskyFeedDefsSkeletonReasonRepost,
				Repost: &repostUri,
			},
		},
	}

	// do request
	ctx := context.Background()
	resp, err := cl.PostAddPostWithResponse(ctx, body)
	if err != nil {
		log.Fatalf("failed to send request: %v", err)
		return
	}

	// check response
	switch resp.StatusCode() {
	case 200:
		fmt.Println("\n=== Response Body ===")
		fmt.Printf("%+v", resp.JSON200)
		return
	case 400, 404, 500:
		log.Fatalf("request error: %s", string(resp.Body))
		return
	default:
		log.Fatalf("unexpected request error: %s", string(resp.Body))
		return
	}
}

type customHeaderTransport struct {
	customHeaders map[string]string
	transport     http.RoundTripper
}

func (c *customHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range c.customHeaders {
		req.Header.Set(key, value)
	}
	if c.transport == nil {
		c.transport = http.DefaultTransport
	}
	return c.transport.RoundTrip(req)
}
