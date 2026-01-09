dev-ui:
	@genkit start -- go run .
.PHONY: dev-ui

debug-notion-mcp:
	@npx @modelcontextprotocol/inspector@latest npx -y @notionhq/notion-mcp-server@latest
.PHONY: debug-notion-mcp

build-recipes:
	@uv run --with streamlit python -m compileall app.py
.PHONY: build-recipes

recipes:
	@uv run --with streamlit streamlit run app.py
.PHONY: recpies