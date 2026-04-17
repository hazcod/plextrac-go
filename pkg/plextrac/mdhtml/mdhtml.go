// Package mdhtml converts a narrow Markdown dialect into the HTML subset
// that Plextrac's Quill-based editor preserves on save.
//
// Plextrac strips style and class attributes from most tags, so this
// converter emits only tags whose default rendering is preserved:
// <p>, <h1>/<h2>/<h3>, <ul>/<ol>/<li>, <strong>, <code>, <pre>, <hr>.
//
// Supported Markdown:
//
//   - ATX headings (#, ##, ###)
//   - Unordered lists: lines matching `-\s+...` or `*\s+...`
//   - Ordered lists: `\d+\.\s+...`
//   - Inline code: `code`
//   - Bold: **bold**
//   - Horizontal rule: --- or *** alone on a line
//   - Fenced code blocks: ``` ... ``` (language ignored, indent preserved)
//
// Everything is HTML-escaped before formatting tags are re-injected. The
// input is pre-decoded via html.UnescapeString so explicit entities round-
// trip instead of double-encoding.
package mdhtml

import (
	"html"
	"regexp"
	"strings"
)

var (
	reInlineCode = regexp.MustCompile("`([^`]+)`")
	reBold       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reHR         = regexp.MustCompile(`^\s*(?:-{3,}|\*{3,})\s*$`)
	reULItem     = regexp.MustCompile(`^\s*[-*]\s+(.+)$`)
	reOLItem     = regexp.MustCompile(`^\s*\d+\.\s+(.+)$`)
)

// Convert renders md as Plextrac-safe HTML.
func Convert(md string) string {
	md = html.UnescapeString(md)
	var out []string
	inCode := false
	var codeBuf []string
	fenceIndent := 0
	inUL := false
	inOL := false
	closeLists := func() {
		if inUL {
			out = append(out, "</ul>")
			inUL = false
		}
		if inOL {
			out = append(out, "</ol>")
			inOL = false
		}
	}

	for _, line := range strings.Split(md, "\n") {
		stripped := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(stripped, "```") {
			if inCode {
				dedented := make([]string, 0, len(codeBuf))
				for _, l := range codeBuf {
					if len(l) >= fenceIndent && strings.TrimLeft(l[:fenceIndent], " \t") == "" {
						dedented = append(dedented, l[fenceIndent:])
					} else {
						dedented = append(dedented, l)
					}
				}
				body := html.EscapeString(strings.Join(dedented, "\n"))
				out = append(out, "<pre><code>"+body+"</code></pre>")
				codeBuf = nil
				inCode = false
				fenceIndent = 0
			} else {
				inCode = true
				fenceIndent = len(line) - len(stripped)
			}
			continue
		}
		if inCode {
			codeBuf = append(codeBuf, line)
			continue
		}
		if reHR.MatchString(line) {
			closeLists()
			out = append(out, "<hr>")
			continue
		}
		if m := reULItem.FindStringSubmatch(line); m != nil {
			if inOL {
				out = append(out, "</ol>")
				inOL = false
			}
			if !inUL {
				out = append(out, "<ul>")
				inUL = true
			}
			out = append(out, "<li>"+inline(m[1])+"</li>")
			continue
		}
		if m := reOLItem.FindStringSubmatch(line); m != nil {
			if inUL {
				out = append(out, "</ul>")
				inUL = false
			}
			if !inOL {
				out = append(out, "<ol>")
				inOL = true
			}
			out = append(out, "<li>"+inline(m[1])+"</li>")
			continue
		}
		closeLists()
		switch {
		case strings.HasPrefix(line, "### "):
			out = append(out, "<h3>"+html.EscapeString(line[4:])+"</h3>")
		case strings.HasPrefix(line, "## "):
			out = append(out, "<h2>"+html.EscapeString(line[3:])+"</h2>")
		case strings.HasPrefix(line, "# "):
			out = append(out, "<h1>"+html.EscapeString(line[2:])+"</h1>")
		case strings.TrimSpace(line) == "":
			out = append(out, "")
		default:
			out = append(out, "<p>"+inline(line)+"</p>")
		}
	}
	closeLists()
	return strings.Join(out, "\n")
}

func inline(s string) string {
	s = html.EscapeString(s)
	s = reInlineCode.ReplaceAllString(s, "<code>$1</code>")
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	return s
}
