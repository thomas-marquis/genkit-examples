package book

import (
	"regexp"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/tmc/langchaingo/textsplitter"
)

var (
	atxHeadingRe      = regexp.MustCompile(`^\s{0,3}#{1,6}(\s+|$).*`)
	setextUnderlineRe = regexp.MustCompile(`^\s{0,3}(=+|-+)\s*$`)
)

func splitDocument(splitter textsplitter.TextSplitter, content string) ([]*ai.Document, error) {
	preparedDocs := make([]*ai.Document, 0)

	chunks, err := splitter.SplitText(content)
	if err != nil {
		return nil, err
	}

	for _, chunk := range chunks {
		if len(chunk) == 0 || onlyContainsHeaders(chunk) { // remove chunks that are only headers
			continue
		}
		chunk = strings.TrimSpace(chunk)
		preparedDocs = append(preparedDocs, ai.DocumentFromText(chunk, nil))
	}

	return preparedDocs, nil
}

func makeTextSplitter() textsplitter.TextSplitter {
	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithKeepSeparator(true),
		textsplitter.WithChunkSize(1000),
		textsplitter.WithHeadingHierarchy(true),
	)
	return splitter
}

// isATXHeading returns true if the given line is an ATX-style heading (#, ##, ..., ######).
func isATXHeading(line string) bool {
	return atxHeadingRe.MatchString(line)
}

// isSetextHeading returns true if the given line is a Setext-style heading (=== or ---).
func isSetextHeading(currLineIdx *int, lines []string) bool {
	j := *currLineIdx + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && setextUnderlineRe.MatchString(strings.TrimSpace(lines[j])) {
		// Treat as Setext heading; skip both lines (and any blanks in between already handled)
		*currLineIdx = j + 1
		return true
	}

	return false
}

func onlyContainsHeaders(text string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		// Whitespace-only is considered as containing no non-heading content.
		return true
	}

	lines := strings.Split(s, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		if isATXHeading(line) {
			i++
			continue
		}
		if isSetextHeading(&i, lines) {
			continue
		}
		return false
	}
	return true
}
