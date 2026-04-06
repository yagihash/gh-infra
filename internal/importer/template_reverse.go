package importer

import (
	"strings"

	"github.com/babarot/gh-infra/internal/fileset"
)

// reverseTemplateContent rebuilds template source from rendered remote content.
// Placeholder-rendered spans are used as anchors; remote literal text around them
// is preserved and placeholder source text is reinserted.
func reverseTemplateContent(templateContent, repo string, vars map[string]string, remote string) (string, bool) {
	trace, err := fileset.RenderTemplateWithTrace(templateContent, repo, vars)
	if err != nil {
		return "", false
	}

	remoteLines := splitKeepNewline(remote)
	var out strings.Builder
	remoteIdx := 0

	for _, line := range trace.Lines {
		if !line.HasPlaceholder {
			continue
		}

		reconstructed, matchedIdx, ok := reverseTemplateLine(line, remoteLines, remoteIdx)
		if !ok {
			return "", false
		}

		for ; remoteIdx < matchedIdx; remoteIdx++ {
			out.WriteString(remoteLines[remoteIdx])
		}
		out.WriteString(reconstructed)
		remoteIdx = matchedIdx + 1
	}

	for ; remoteIdx < len(remoteLines); remoteIdx++ {
		out.WriteString(remoteLines[remoteIdx])
	}

	return out.String(), true
}

func reverseTemplateLine(line fileset.RenderedLine, remoteLines []string, start int) (string, int, bool) {
	for idx := start; idx < len(remoteLines); idx++ {
		if reconstructed, ok := reconstructLineFromRemote(line, remoteLines[idx]); ok {
			return reconstructed, idx, true
		}
	}
	return "", -1, false
}

func reconstructLineFromRemote(line fileset.RenderedLine, remote string) (string, bool) {
	var out strings.Builder
	pos := 0

	for _, seg := range line.Segments {
		if seg.Kind != fileset.SegmentPlaceholder {
			continue
		}

		idx := strings.Index(remote[pos:], seg.RenderedText)
		if idx < 0 {
			return "", false
		}
		idx += pos
		out.WriteString(remote[pos:idx])
		out.WriteString(seg.SourceText)
		pos = idx + len(seg.RenderedText)
	}

	out.WriteString(remote[pos:])
	return out.String(), true
}

func supportsTemplateReverse(templateContent, repo string, vars map[string]string) bool {
	_, err := fileset.RenderTemplateWithTrace(templateContent, repo, vars)
	return err == nil
}

func splitKeepNewline(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
