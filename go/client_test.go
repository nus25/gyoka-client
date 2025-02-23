package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/nus25/gyoka-client/go/types"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	logger := slog.Default()

	t.Run("CreatePosts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/feed/add", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			var req CreatePostRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Len(t, req.Posts, 1)

			resp := CreatePostResponse{
				InsertedPosts: req.Posts,
				Message:       "success",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(server.URL, logger)
		assert.NoError(t, err)

		ctx := context.Background()
		posts := []types.Post{
			{
				Feed: "at://did:plc:test/app.bsky.feed.generator/test",
				Uri:  "at://did:plc:test/app.bsky.feed.post/test",
			},
		}

		resp, err := client.Add(ctx, posts)
		assert.NoError(t, err)
		assert.Len(t, resp.InsertedPosts, 1)
		assert.Equal(t, posts[0].Feed, resp.InsertedPosts[0].Feed)
		assert.Equal(t, posts[0].Uri, resp.InsertedPosts[0].Uri)
		assert.Equal(t, "success", resp.Message)
	})

	t.Run("DeletePosts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/feed/delete", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			var req DeletePostRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Len(t, req.Posts, 1)

			resp := DeletePostResponse{
				DeletedPosts: req.Posts,
				Message:      "success",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(server.URL, logger)
		assert.NoError(t, err)

		ctx := context.Background()
		posts := []types.Post{
			{
				Feed: "at://did:plc:test/app.bsky.feed.generator/test",
				Uri:  "at://did:plc:test/app.bsky.feed.post/test",
			},
		}

		resp, err := client.Delete(ctx, posts)
		assert.Len(t, resp.DeletedPosts, 1)
		assert.Equal(t, posts[0].Feed, resp.DeletedPosts[0].Feed)
		assert.Equal(t, posts[0].Uri, resp.DeletedPosts[0].Uri)
		assert.Equal(t, "success", resp.Message)
		assert.NoError(t, err)
	})

	t.Run("Auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreatePostResponse{})
		}))
		defer server.Close()

		client, err := NewClient(server.URL, logger, WithToken("test-token"))
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = client.Add(ctx, []types.Post{})
		assert.NoError(t, err)
	})

	t.Run("RetryOnError", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/feed/add", r.URL.Path) // パスを明示的に確認
			assert.Equal(t, "POST", r.Method)        // メソッドも確認

			attempts++
			t.Logf("Attempt %d received", attempts) // デバッグ用のログ追加
			if attempts < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreatePostResponse{})
		}))
		defer server.Close()

		client, err := NewClient(server.URL, logger)
		assert.NoError(t, err)

		ctx := context.Background()
		posts := []types.Post{
			{
				Feed: "at://did:plc:test/app.bsky.feed.generator/test",
				Uri:  "at://did:plc:test/app.bsky.feed.post/test1",
				Cid:  "test-cid-1",
			},
			{
				Feed: "at://did:plc:test/app.bsky.feed.generator/test",
				Uri:  "at://did:plc:test/app.bsky.feed.post/test2",
				Cid:  "test-cid-2",
			},
		}
		_, err = client.Add(ctx, posts)
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})
}
