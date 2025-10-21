package agent

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

func setupMCPs(ctx context.Context, g *genkit.Genkit, notionKey, todoistKey string) ([]ai.Tool, error) {
	mcpOpts := []mcp.MCPClientOptions{
		{
			Name: "todoist-mcp-server",
			StreamableHTTP: &mcp.StreamableHTTPConfig{
				BaseURL: "https://ai.todoist.net/mcp",
				//Headers: map[string]string{
				//	"Authorization": fmt.Sprintf("Bearer %s", todoistKey),
				//},
			},
		},
		{
			Name: "notion-mcp-server",
			Stdio: &mcp.StdioConfig{
				Command: "npx",
				Args:    []string{"-y", "@notionhq/notion-mcp-server"},
				Env:     []string{fmt.Sprintf("NOTION_TOKEN=%s", notionKey)},
			},
		},
	}

	var tools []ai.Tool
	for _, opt := range mcpOpts {
		client, err := mcp.NewGenkitMCPClient(opt)
		if err != nil {
			return nil, err
		}
		at, err := client.GetActiveTools(ctx, g)
		if err != nil {
			return nil, fmt.Errorf("error getting tools from %s: %w", opt.Name, err)
		}
		tools = append(tools, at...)
	}

	for _, t := range tools {
		genkit.RegisterAction(g, t)
	}

	return tools, nil
}
