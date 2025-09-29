package agent

import (
	"fmt"

	"github.com/firebase/genkit/go/plugins/mcp"
)

func defineNotionMcp(secretKey string) (*mcp.GenkitMCPClient, error) {
	client, err := mcp.NewGenkitMCPClient(mcp.MCPClientOptions{
		Name: "notion-mcp-server",
		Stdio: &mcp.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "@notionhq/notion-mcp-server"},
			Env:     []string{fmt.Sprintf("NOTION_TOKEN=%s", secretKey)},
		},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
