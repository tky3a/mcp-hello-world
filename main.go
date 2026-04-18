package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// 1. クライアントの作成（NewStdioMCPClient を使用）
	// server-filesystem は、引数としてアクセスを許可するディレクトリのパスを受け取ります。
	c, err := client.NewStdioMCPClient(
		"npx",
		os.Environ(),
		"-y", "@modelcontextprotocol/server-filesystem",
		"/Users/tky/workspace/WebApplication/mcp-hello-world",
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 3. 初期化（Initialize）の実行
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "my-go-client",
				Version: "1.0.0",
			},
		},
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Initialize error: %v", err)
	}

	fmt.Println("MCP Client Initialized successfully!")

	// 4. 利用可能なツール一覧を取得して表示
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("ListTools error: %v", err)
	}

	fmt.Println("\n--- Available Tools ---")
	for _, t := range tools.Tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// 5. 特定のツールを実行（read_file）
	fmt.Println("\n--- Executing Tool: read_file ---")
	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "read_file",
			Arguments: map[string]any{
				"path": "/Users/tky/workspace/WebApplication/mcp-hello-world/test_memo.txt",
			},
		},
	})

	if err != nil {
		log.Fatalf("CallTool error: %v", err)
	}

	fmt.Printf("Result: %v\n", result.Content)

}
