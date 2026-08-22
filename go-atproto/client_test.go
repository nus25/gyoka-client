package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lexutil "github.com/bluesky-social/indigo/lex/util"
)

func TestNewWithDirectoryAuthenticatesAndProxiesRequests(t *testing.T) {
	const (
		identifier  = "sample.bsky.social"
		appPassword = "test-app-password"
		accessToken = "test-access-token"
		did         = "did:plc:ewvi7nxzyoun6zhxrhs64oiz"
		host        = "editor.example.com"
	)

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/xrpc/com.atproto.server.createSession":
			if request.Method != http.MethodPost {
				t.Errorf("login method = %s, want %s", request.Method, http.MethodPost)
			}

			var input struct {
				Identifier string `json:"identifier"`
				Password   string `json:"password"`
			}
			if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
				t.Fatalf("decode login request: %v", err)
			}
			if input.Identifier != did {
				t.Errorf("login identifier = %q, want %q", input.Identifier, did)
			}
			if input.Password != appPassword {
				t.Errorf("login password = %q, want app password", input.Password)
			}

			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{"did":"did:plc:ewvi7nxzyoun6zhxrhs64oiz","accessJwt":"test-access-token","refreshJwt":"test-refresh-token"}`))
		case "/xrpc/net.nusno.gyoka.ping":
			if request.Method != http.MethodGet {
				t.Errorf("ping method = %s, want %s", request.Method, http.MethodGet)
			}
			if authorization := request.Header.Get("Authorization"); authorization != "Bearer "+accessToken {
				t.Errorf("Authorization = %q, want bearer access token", authorization)
			}
			if proxy := request.Header.Get("Atproto-Proxy"); proxy != "did:web:"+host+"#gyoka_editor" {
				t.Errorf("Atproto-Proxy = %q, want Gyoka service reference", proxy)
			}

			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{"message":"ok"}`))
		default:
			t.Errorf("unexpected request path: %s", request.URL.Path)
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	directory := identity.NewMockDirectory()
	directory.Insert(identity.Identity{
		DID:    syntax.DID(did),
		Handle: syntax.Handle(identifier),
		Services: map[string]identity.ServiceEndpoint{
			"atproto_pds": {URL: server.URL},
		},
	})

	client, err := newWithDirectory(context.Background(), host, identifier, appPassword, directory)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	output, err := client.Ping(context.Background())
	if err != nil {
		t.Fatalf("ping: %v", err)
	}
	if output.Message != "ok" {
		t.Errorf("ping message = %q, want %q", output.Message, "ok")
	}
}

func TestClientDelegatesEveryLexiconEndpoint(t *testing.T) {
	recorder := &recordingLexClient{}
	client := &Client{lexClient: recorder}
	ctx := context.Background()

	_, _ = client.Ping(ctx)
	_, _ = client.AddPost(ctx, nil)
	_, _ = client.BatchAddPosts(ctx, nil)
	_, _ = client.BatchRemovePosts(ctx, nil)
	_, _ = client.GetPosts(ctx, "cursor", "at://did:plc:test/app.bsky.feed.generator/feed", 10)
	_, _ = client.ListFeeds(ctx)
	_, _ = client.RegisterFeed(ctx, nil)
	_, _ = client.RemovePost(ctx, nil)
	_, _ = client.RemovePostByAuthor(ctx, nil)
	_, _ = client.TrimFeed(ctx, nil)
	_, _ = client.UnregisterFeed(ctx, nil)
	_, _ = client.UpdateFeed(ctx, nil)
	_, _ = client.GetDocument(ctx, "about")
	_, _ = client.UpdateDocument(ctx, nil)

	want := []string{
		"net.nusno.gyoka.ping",
		"net.nusno.gyoka.feed.addPost",
		"net.nusno.gyoka.feed.batchAddPosts",
		"net.nusno.gyoka.feed.batchRemovePosts",
		"net.nusno.gyoka.feed.getPosts",
		"net.nusno.gyoka.feed.listFeeds",
		"net.nusno.gyoka.feed.registerFeed",
		"net.nusno.gyoka.feed.removePost",
		"net.nusno.gyoka.feed.removePostByAuthor",
		"net.nusno.gyoka.feed.trimFeed",
		"net.nusno.gyoka.feed.unregisterFeed",
		"net.nusno.gyoka.feed.updateFeed",
		"net.nusno.gyoka.document.getDocument",
		"net.nusno.gyoka.document.updateDocument",
	}
	if !reflect.DeepEqual(recorder.endpoints, want) {
		t.Errorf("endpoints = %#v, want %#v", recorder.endpoints, want)
	}
}

func TestPingPropagatesLexiconError(t *testing.T) {
	want := errors.New("service unavailable")
	client := &Client{lexClient: &recordingLexClient{err: want}}

	_, err := client.Ping(context.Background())
	if !errors.Is(err, want) {
		t.Errorf("ping error = %v, want %v", err, want)
	}
}

type recordingLexClient struct {
	endpoints []string
	err       error
}

var _ lexutil.LexClient = (*recordingLexClient)(nil)

func (c *recordingLexClient) LexDo(_ context.Context, _ string, _ string, endpoint string, _ map[string]any, _ any, _ any) error {
	c.endpoints = append(c.endpoints, endpoint)
	return c.err
}
