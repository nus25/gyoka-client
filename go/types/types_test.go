package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedUriValidate(t *testing.T) {
	tests := []struct {
		name    string
		uri     FeedUri
		wantErr bool
	}{
		{
			name:    "valid feed uri",
			uri:     "at://did:plc:1234/app.bsky.feed.generator/test",
			wantErr: false,
		},
		{
			name:    "empty uri",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			uri:     "invalid://did:plc:1234/app.bsky.feed.generator/test",
			wantErr: true,
		},
		{
			name:    "invalid did",
			uri:     "at://invalid/app.bsky.feed.generator/test",
			wantErr: true,
		},
		{
			name:    "invalid collection",
			uri:     "at://did:plc:1234/invalid/test",
			wantErr: true,
		},
		{
			name:    "empty feed name",
			uri:     "at://did:plc:1234/app.bsky.feed.generator/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.uri.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostUriValidate(t *testing.T) {
	tests := []struct {
		name    string
		uri     PostUri
		wantErr bool
	}{
		{
			name:    "valid post uri",
			uri:     "at://did:plc:1234/app.bsky.feed.post/test",
			wantErr: false,
		},
		{
			name:    "empty uri",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			uri:     "invalid://did:plc:1234/app.bsky.feed.post/test",
			wantErr: true,
		},
		{
			name:    "invalid did",
			uri:     "at://invalid/app.bsky.feed.post/test",
			wantErr: true,
		},
		{
			name:    "invalid collection",
			uri:     "at://did:plc:1234/invalid/test",
			wantErr: true,
		},
		{
			name:    "empty post id",
			uri:     "at://did:plc:1234/app.bsky.feed.post/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.uri.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
