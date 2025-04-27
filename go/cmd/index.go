package main

//Sample usage of this client.

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	client "github.com/nus25/gyoka-client/go"
)

func main() {
	client, err := client.NewClient("http://localhost:8787", client.WithHTTPClient(&http.Client{}))
	if err != nil {
		log.Fatalf("クライアントの作成に失敗しました: %v", err)
	}

	resp, err := client.GetListFeed(context.Background())
	if err != nil {
		log.Fatalf("API 呼び出しに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("\n=== Response Body ===")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("レスポンスボディの読み取りに失敗しました: %v", err)
	}
	fmt.Println(string(bodyBytes))
}
