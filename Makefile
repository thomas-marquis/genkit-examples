dev-ui:
	@genkit start -- go run .
.PHONY: dev-ui


debug-mcp:
	@npx @modelcontextprotocol/inspector npx -y @notionhq/notion-mcp-server
.PHONY: debug-mcp