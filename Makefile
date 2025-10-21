dev-ui:
	@genkit start -- go run .
.PHONY: dev-ui


debug-notion-mcp:
	@npx @modelcontextprotocol/inspector npx -y @notionhq/notion-mcp-server
.PHONY: debug-notion-mcp

debug-todoist-mcp:
	@npx @modelcontextprotocol/inspector npx -y mcp-remote https://ai.todoist.net/mcp
.PHONY: debug-notion-mcp