package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/atcrypto"
	"github.com/bluesky-social/indigo/atproto/auth"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	gyoka "github.com/nus25/gyoka-client/go-atproto/schema/gyoka"
)

const (
	gyokaServiceFragment = "gyoka_editor"
	serviceAuthTTL       = time.Minute
)

// Client calls Gyoka Lexicon endpoints through an authenticated PDS service proxy.
type Client struct {
	lexClient lexutil.LexClient
}

// InterServiceAuthConfig configures direct authentication to a Gyoka service.
type InterServiceAuthConfig struct {
	Host       string
	Audience   string
	Issuer     syntax.DID
	PrivateKey atcrypto.PrivateKey
}

type serviceAuthMethod struct {
	audience   string
	issuer     syntax.DID
	privateKey atcrypto.PrivateKey
}

func (a *serviceAuthMethod) DoWithAuth(httpClient *http.Client, request *http.Request, endpoint syntax.NSID) (*http.Response, error) {
	token, err := auth.SignServiceAuth(a.issuer, a.audience, serviceAuthTTL, &endpoint, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign inter-service authentication JWT: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	return httpClient.Do(request)
}

// New creates a Client for the Gyoka service at host.
//
// It resolves identifier to its PDS, then authenticates there with appPassword.
// The resulting session remains in memory and is used for requests proxied to
// did:web:<host>#gyoka_editor.
func New(ctx context.Context, host, identifier, appPassword string) (*Client, error) {
	return newWithDirectory(ctx, host, identifier, appPassword, identity.DefaultDirectory())
}

func newWithDirectory(ctx context.Context, host, identifier, appPassword string, directory identity.Directory) (*Client, error) {
	atIdentifier, err := syntax.ParseAtIdentifier(identifier)
	if err != nil {
		return nil, fmt.Errorf("parse identifier: %w", err)
	}

	apiClient, err := atclient.LoginWithPassword(ctx, directory, atIdentifier, appPassword, "", nil)
	if err != nil {
		return nil, fmt.Errorf("log in with app password: %w", normalizeAPIError(err))
	}

	return NewWithAPIClient(host, apiClient)
}

// NewWithInterServiceAuth creates a Client that connects directly to Host.
//
// Audience is the full receiving service reference used in the JWT aud claim.
// Issuer and PrivateKey must correspond to the issuer DID's #atproto key.
func NewWithInterServiceAuth(config InterServiceAuthConfig) (*Client, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("Gyoka host is required")
	}
	if config.Audience == "" {
		return nil, fmt.Errorf("Gyoka audience is required")
	}
	if config.Issuer == "" {
		return nil, fmt.Errorf("issuer DID is required")
	}
	if config.PrivateKey == nil {
		return nil, fmt.Errorf("issuer private key is required")
	}

	apiClient := atclient.NewAPIClient(config.Host)
	apiClient.Auth = &serviceAuthMethod{
		audience:   config.Audience,
		issuer:     config.Issuer,
		privateKey: config.PrivateKey,
	}
	return &Client{lexClient: apiClient}, nil
}

// NewWithAPIClient creates a Client using an authenticated Indigo API client.
//
// It configures apiClient to proxy requests to did:web:<host>#gyoka_editor.
// Use it to provide a client authenticated through a mechanism other than an
// app password.
func NewWithAPIClient(host string, apiClient *atclient.APIClient) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("Gyoka host is required")
	}
	if apiClient == nil {
		return nil, fmt.Errorf("API client is required")
	}

	serviceRef := "did:web:" + host + "#" + gyokaServiceFragment
	return &Client{lexClient: apiClient.WithService(serviceRef)}, nil
}

func (c *Client) Ping(ctx context.Context) (*gyoka.Ping_Output, error) {
	return normalizeResult(gyoka.Ping(ctx, c.lexClient))
}

func (c *Client) RegisterFeed(ctx context.Context, input *gyoka.FeedRegisterFeed_Input) (*gyoka.FeedRegisterFeed_Output, error) {
	return normalizeResult(gyoka.FeedRegisterFeed(ctx, c.lexClient, input))
}

func (c *Client) ListFeeds(ctx context.Context) (*gyoka.FeedListFeeds_Output, error) {
	return normalizeResult(gyoka.FeedListFeeds(ctx, c.lexClient))
}

func (c *Client) UpdateFeed(ctx context.Context, input *gyoka.FeedUpdateFeed_Input) (*gyoka.FeedUpdateFeed_Output, error) {
	return normalizeResult(gyoka.FeedUpdateFeed(ctx, c.lexClient, input))
}

func (c *Client) UnregisterFeed(ctx context.Context, input *gyoka.FeedUnregisterFeed_Input) (*gyoka.FeedUnregisterFeed_Output, error) {
	return normalizeResult(gyoka.FeedUnregisterFeed(ctx, c.lexClient, input))
}

func (c *Client) PutPosts(ctx context.Context, input *gyoka.FeedPutPosts_Input) (*gyoka.FeedPutPosts_Output, error) {
	return normalizeResult(gyoka.FeedPutPosts(ctx, c.lexClient, input))
}

func (c *Client) RemovePosts(ctx context.Context, input *gyoka.FeedRemovePosts_Input) (*gyoka.FeedRemovePosts_Output, error) {
	return normalizeResult(gyoka.FeedRemovePosts(ctx, c.lexClient, input))
}

func (c *Client) RemovePostsByAuthor(ctx context.Context, input *gyoka.FeedRemovePostsByAuthor_Input) (*gyoka.FeedRemovePostsByAuthor_Output, error) {
	return normalizeResult(gyoka.FeedRemovePostsByAuthor(ctx, c.lexClient, input))
}

func (c *Client) MovePosts(ctx context.Context, input *gyoka.FeedMovePosts_Input) (*gyoka.FeedMovePosts_Output, error) {
	return normalizeResult(gyoka.FeedMovePosts(ctx, c.lexClient, input))
}

func (c *Client) GetPosts(ctx context.Context, feed, postUri, author string, limit int64, cursor string) (*gyoka.FeedGetPosts_Output, error) {
	// sort arguments as required by the API: author, cursor, feed, limit, postUri
	return normalizeResult(gyoka.FeedGetPosts(ctx, c.lexClient, author, cursor, feed, limit, postUri))
}

func (c *Client) TrimPosts(ctx context.Context, input *gyoka.FeedTrimPosts_Input) (*gyoka.FeedTrimPosts_Output, error) {
	return normalizeResult(gyoka.FeedTrimPosts(ctx, c.lexClient, input))
}

func (c *Client) PrunePostsBefore(ctx context.Context, input *gyoka.FeedPrunePostsBefore_Input) (*gyoka.FeedPrunePostsBefore_Output, error) {
	return normalizeResult(gyoka.FeedPrunePostsBefore(ctx, c.lexClient, input))
}

func (c *Client) RebalanceOrder(ctx context.Context, input *gyoka.FeedRebalanceOrder_Input) (*gyoka.FeedRebalanceOrder_Output, error) {
	return normalizeResult(gyoka.FeedRebalanceOrder(ctx, c.lexClient, input))
}

func (c *Client) GetDocument(ctx context.Context, docType string) (*gyoka.DocumentGetDocument_Output, error) {
	return normalizeResult(gyoka.DocumentGetDocument(ctx, c.lexClient, docType))
}

func (c *Client) UpdateDocument(ctx context.Context, input *gyoka.DocumentUpdateDocument_Input) (*gyoka.DocumentUpdateDocument_Output, error) {
	return normalizeResult(gyoka.DocumentUpdateDocument(ctx, c.lexClient, input))
}
