package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// authTransport はすべてのリクエストに API Key ヘッダーを付与するカスタム RoundTripper です
type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Api-Key", t.apiKey)
	req.Header.Set("include-tags", "discovery,alerting")
	req.Header.Set("Content-Type", "application/json")

	// リクエスト直前にデバッグログ
	fmt.Printf("Connecting to: %s with header Api-Key: %s...\n", req.URL, t.apiKey[:10]+"***")

	resp, err := t.base.RoundTrip(req)

	// 403が出た場合に詳細を表示
	if resp != nil && resp.StatusCode == 403 {
		fmt.Println("Error: 403 Forbidden. Please check if your User API Key (NRAK-...) is correct and has sufficient permissions.")
	}
	return resp, err
}

func main() {
	_ = godotenv.Load()
	apiKey := os.Getenv("NEW_RELIC_API_KEY")
	if apiKey == "" {
		log.Fatal("NEW_RELIC_API_KEY is not set")
	}

	ctx := context.Background()

	// 1. カスタム HTTP クライアントの作成
	// これにより、SSE の接続確立時やその後の通信すべてに API キーが乗ります
	httpClient := &http.Client{
		Transport: &authTransport{
			apiKey: apiKey,
			base:   http.DefaultTransport,
		},
	}

	endpoint := "https://mcp.newrelic.com/mcp/"

	// 2. HTTP クライアントを渡して SSE クライアントを生成
	c, err := client.NewSSEMCPClient(
		endpoint,
		client.WithHTTPClient(httpClient), // ここで作成したクライアントを渡す
	)
	if err != nil {
		log.Fatalf("Failed to create SSE client: %v", err)
	}
	defer c.Close()

	// --- ここが重要！ ---
	// SSEの場合は、Initializeの前にStartを呼んで接続を確立させる必要があります
	if err := c.Start(ctx); err != nil {
		log.Fatalf("Failed to start transport: %v", err)
	}
	// -------------------

	// 3. 初期化
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "my-go-mcp-hub",
				Version: "1.0.0",
			},
		},
	}

	// 認証が通っていれば、ここで成功します
	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Initialize error: %v", err)
	}

	fmt.Println("Successfully connected to New Relic MCP via SSE!")
}
