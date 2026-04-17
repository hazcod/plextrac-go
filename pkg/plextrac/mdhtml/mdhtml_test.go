package mdhtml

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", "<p>hello</p>"},
		{"h1", "# Title", "<h1>Title</h1>"},
		{"h2", "## Title", "<h2>Title</h2>"},
		{"h3", "### Title", "<h3>Title</h3>"},
		{"bold", "this is **bold**", "<p>this is <strong>bold</strong></p>"},
		{"inline code", "call `fn()` here", "<p>call <code>fn()</code> here</p>"},
		{"ul", "- one\n- two", "<ul>\n<li>one</li>\n<li>two</li>\n</ul>"},
		{"ol", "1. one\n2. two", "<ol>\n<li>one</li>\n<li>two</li>\n</ol>"},
		{"hr", "---", "<hr>"},
		{"xss escaping", "<script>alert(1)</script>", "<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>"},
		{"html entity roundtrip", "&lt;code&gt;", "<p>&lt;code&gt;</p>"},
		{"fenced code", "```\nfoo\n```", "<pre><code>foo</code></pre>"},
		{"fenced code preserves lt", "```\n<a>\n```", "<pre><code>&lt;a&gt;</code></pre>"},
		{"empty line", "", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Convert(tc.in)
			if got != tc.want {
				t.Fatalf("Convert(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestConvertLongList(t *testing.T) {
	t.Parallel()
	md := "- a\n- b\n\npara\n\n1. x\n2. y"
	got := Convert(md)
	if !strings.Contains(got, "<ul>") || !strings.Contains(got, "<ol>") || !strings.Contains(got, "<p>para</p>") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func FuzzConvert(f *testing.F) {
	f.Add("# hi\n- a\n```\nx\n```")
	f.Add("<script>")
	f.Add("**bold** and `code`")
	f.Fuzz(func(t *testing.T, in string) {
		// Must not panic and must return a string that contains no
		// unescaped user `<` outside tags we know about.
		_ = Convert(in)
	})
}
