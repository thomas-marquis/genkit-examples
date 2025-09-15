package book

import (
	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"
)

func makeMarkdownConverter() *converter.Converter {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	conv.Register.PreRenderer(fixLinksPreRender, converter.PriorityEarly) // fix HTML self-closed <a/> rendering
	return conv
}

func fixLinksPreRender(ctx converter.Context, doc *html.Node) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := dom.GetAttributeOr(n, "href", "")
			hasChildren := n.FirstChild != nil

			if (href == "" || href == "#") && hasChildren {
				n.Data = "div"
				n.Attr = nil
			} else if !hasChildren && href != "" && href != "#" {
				textNode := &html.Node{
					Type: html.TextNode,
					Data: href,
				}
				n.AppendChild(textNode)
			} else {
				n.Data = "p"
				n.Attr = nil
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
}
