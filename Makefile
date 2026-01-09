dev-ui:
	@genkit start -- go run .
.PHONY: dev-ui

debug-notion-mcp:
	@npx @modelcontextprotocol/inspector@latest npx -y @notionhq/notion-mcp-server@latest
.PHONY: debug-notion-mcp
