package client

import (
	"context"
	"fmt"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	gyoka "github.com/nus25/gyoka-client/go-atproto/schema/gyoka"
)

const gyokaServiceFragment = "gyoka_editor"

// Client calls Gyoka Lexicon endpoints through an authenticated PDS service proxy.
type Client struct {
	lexClient lexutil.LexClient
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
		return nil, fmt.Errorf("log in with app password: %w", err)
	}

	return NewWithAPIClient(host, apiClient)
}

// NewWithAPIClient creates a Client using an authenticated Indigo API client.
//
// It configures apiClient to proxy requests to did:web:<host>#gyoka_editor.
// Use it to provide a client authenticated through a mechanism other than an
// app password, such as future inter-service authentication.
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
	return gyoka.Ping(ctx, c.lexClient)
}

func (c *Client) AddPost(ctx context.Context, input *gyoka.FeedAddPost_Input) (*gyoka.FeedAddPost_Output, error) {
	return gyoka.FeedAddPost(ctx, c.lexClient, input)
}

func (c *Client) BatchAddPosts(ctx context.Context, input *gyoka.FeedBatchAddPosts_Input) (*gyoka.FeedBatchAddPosts_Output, error) {
	return gyoka.FeedBatchAddPosts(ctx, c.lexClient, input)
}

func (c *Client) BatchRemovePosts(ctx context.Context, input *gyoka.FeedBatchRemovePosts_Input) (*gyoka.FeedBatchRemovePosts_Output, error) {
	return gyoka.FeedBatchRemovePosts(ctx, c.lexClient, input)
}

func (c *Client) GetPosts(ctx context.Context, cursor, feed string, limit int64) (*gyoka.FeedGetPosts_Output, error) {
	return gyoka.FeedGetPosts(ctx, c.lexClient, cursor, feed, limit)
}

func (c *Client) ListFeeds(ctx context.Context) (*gyoka.FeedListFeeds_Output, error) {
	return gyoka.FeedListFeeds(ctx, c.lexClient)
}

func (c *Client) RegisterFeed(ctx context.Context, input *gyoka.FeedRegisterFeed_Input) (*gyoka.FeedRegisterFeed_Output, error) {
	return gyoka.FeedRegisterFeed(ctx, c.lexClient, input)
}

func (c *Client) RemovePost(ctx context.Context, input *gyoka.FeedRemovePost_Input) (*gyoka.FeedRemovePost_Output, error) {
	return gyoka.FeedRemovePost(ctx, c.lexClient, input)
}

func (c *Client) RemovePostByAuthor(ctx context.Context, input *gyoka.FeedRemovePostByAuthor_Input) (*gyoka.FeedRemovePostByAuthor_Output, error) {
	return gyoka.FeedRemovePostByAuthor(ctx, c.lexClient, input)
}

func (c *Client) TrimFeed(ctx context.Context, input *gyoka.FeedTrimFeed_Input) (*gyoka.FeedTrimFeed_Output, error) {
	return gyoka.FeedTrimFeed(ctx, c.lexClient, input)
}

func (c *Client) UnregisterFeed(ctx context.Context, input *gyoka.FeedUnregisterFeed_Input) (*gyoka.FeedUnregisterFeed_Output, error) {
	return gyoka.FeedUnregisterFeed(ctx, c.lexClient, input)
}

func (c *Client) UpdateFeed(ctx context.Context, input *gyoka.FeedUpdateFeed_Input) (*gyoka.FeedUpdateFeed_Output, error) {
	return gyoka.FeedUpdateFeed(ctx, c.lexClient, input)
}

func (c *Client) GetDocument(ctx context.Context, docType string) (*gyoka.DocumentGetDocument_Output, error) {
	return gyoka.DocumentGetDocument(ctx, c.lexClient, docType)
}

func (c *Client) UpdateDocument(ctx context.Context, input *gyoka.DocumentUpdateDocument_Input) (*gyoka.DocumentUpdateDocument_Output, error) {
	return gyoka.DocumentUpdateDocument(ctx, c.lexClient, input)
}
