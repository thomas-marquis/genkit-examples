package domain

import "strings"

type Recipe struct {
	Title       string
	Ingredients []string
	Steps       []string
}

func (r Recipe) String() string {
	sb := strings.Builder{}
	sb.WriteString("# ")
	sb.WriteString(r.Title)
	sb.WriteString("\n")
	sb.WriteString("\n## Ingredients:\n")
	for _, ing := range r.Ingredients {
		sb.WriteString(ing)
		sb.WriteString("\n")
	}
	sb.WriteString("\n## Steps:\n")
	for _, step := range r.Steps {
		sb.WriteString(step)
		sb.WriteString("\n")
	}
	return sb.String()
}
