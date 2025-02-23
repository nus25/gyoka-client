package types

import (
	"errors"
	"strings"
)

type Post struct {
	Feed      FeedUri `json:"feed,omitempty"`
	Uri       PostUri `json:"uri"`
	Cid       string  `json:"cid"`
	IndexedAt string  `json:"indexedAt"`
}

type AtUri string

func (u AtUri) validate(collection string, idName string) error {
	if u == "" {
		return errors.New("uri is empty")
	}

	s := string(u)
	if !strings.HasPrefix(s, "at://") {
		return errors.New("uri must start with at://")
	}

	parts := strings.Split(s[5:], "/")
	if len(parts) != 3 {
		return errors.New("invalid uri format")
	}

	if !strings.HasPrefix(parts[0], "did:plc:") {
		return errors.New("invalid did format")
	}

	if parts[1] != collection {
		return errors.New("invalid collection")
	}

	if parts[2] == "" {
		return errors.New(idName + " is empty")
	}

	return nil
}

type FeedUri string

func (f FeedUri) Validate() error {
	return AtUri(f).validate("app.bsky.feed.generator", "feed name")
}

type PostUri string

func (u PostUri) Validate() error {
	return AtUri(u).validate("app.bsky.feed.post", "post id")
}
