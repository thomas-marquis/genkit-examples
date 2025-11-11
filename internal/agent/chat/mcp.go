package chat

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

func (f *Flow) setupMCPClient(ctx context.Context) error {
	if f.notionApiKey == "" {
		logger.Println("Notion API key not provided, skipping MCP setup")
		return nil
	}

	notionMcpOpts := mcp.MCPClientOptions{
		Name: "notion-mcp-server",
		Stdio: &mcp.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "@notionhq/notion-mcp-server"},
			Env:     []string{fmt.Sprintf("NOTION_TOKEN=%s", f.notionApiKey)},
		},
	}

	client, err := mcp.NewGenkitMCPClient(notionMcpOpts)
	if err != nil {
		return err
	}
	tools, err := client.GetActiveTools(ctx, f.g)
	if err != nil {
		return fmt.Errorf("error getting tools from %s: %w", notionMcpOpts.Name, err)
	}

	for _, t := range tools {
		genkit.RegisterAction(f.g, t)
	}

	return nil
}
