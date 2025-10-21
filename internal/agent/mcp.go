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
			Name: "notion-mcp-server",
			Stdio: &mcp.StdioConfig{
				Command: "npx",
				Args:    []string{"-y", "@notionhq/notion-mcp-server"},
				Env:     []string{fmt.Sprintf("NOTION_TOKEN=%s", notionKey)},
			},
		},
		{
			Name: "todoist-mcp-server",
			Stdio: &mcp.StdioConfig{
				Command: "npx",
				Args:    []string{"-y", "@doist/todoist-ai"},
				Env:     []string{fmt.Sprintf("TODOIST_API_KEY=%s", todoistKey)},
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

	return tools, nil
}
