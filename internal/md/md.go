// Package md parses the Markdown subset of the content files into a small block tree. The views render that tree with
// UI8Kit 3 primitives only (no HTML string is ever produced or injected), and the same source files are served raw to
// agents (/<lang>/<page>.md, llms.txt, llms-full.txt).
//
// Supported: YAML-like front matter (key: value), ATX headings (# .. ####), paragraphs, fenced code (```lang),
// bullet and numbered lists (one level, lazy continuation lines), GFM tables, blockquotes (rendered as a note), "---"
// rules. Inline: `code`, **strong**, *em*, [text](href). Anything else stays plain text.
package md

import (
	"regexp"
	"strings"
)

// Doc is one parsed page.
type Doc struct {
	Meta   map[string]string
	Blocks []Block
	Source string // the file as written (front matter included)
	Body   string // the file without front matter
}

// Kinds of blocks.
const (
	Heading   = "heading"
	Paragraph = "paragraph"
	Code      = "code"
	Bullets   = "bullets"
	Numbers   = "numbers"
	Table     = "table"
	Note      = "note"
	Rule      = "rule"
)

// Block is one block of a page.
type Block struct {
	Kind   string
	Level  int        // heading level
	ID     string     // heading anchor
	Text   []Inline   // heading, paragraph
	Lang   string     // code language
	Code   string     // code text
	Items  [][]Inline // list items, note paragraphs
	Header [][]Inline // table header cells
	Rows   [][][]Inline
}

// Kinds of inline spans.
const (
	Text   = "text"
	Mono   = "code"
	Strong = "strong"
	Em     = "em"
	Link   = "link"
)

// Inline is a span of text.
type Inline struct {
	Kind string
	Text string
	Href string
}

var (
	fence    = regexp.MustCompile("^```\\s*([A-Za-z0-9_+-]*)\\s*$")
	heading  = regexp.MustCompile(`^(#{1,4})\s+(.+?)\s*#*\s*$`)
	bullet   = regexp.MustCompile(`^[-*]\s+(.*)$`)
	number   = regexp.MustCompile(`^\d+[.)]\s+(.*)$`)
	tableSep = regexp.MustCompile(`^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$`)
	inlineRe = regexp.MustCompile("`([^`]+)`|\\*\\*([^*]+)\\*\\*|\\*([^*\\s][^*]*)\\*|\\[([^\\]]+)\\]\\(([^)\\s]+)\\)")
	slugDrop = regexp.MustCompile(`[^\p{L}\p{N}]+`)
)

// Parse parses one file.
func Parse(src string) Doc {
	d := Doc{Meta: map[string]string{}, Source: src}
	body := strings.ReplaceAll(src, "\r\n", "\n")
	if strings.HasPrefix(body, "---\n") {
		if end := strings.Index(body[4:], "\n---\n"); end >= 0 {
			for _, l := range strings.Split(body[4:4+end], "\n") {
				if k, v, ok := strings.Cut(l, ":"); ok {
					d.Meta[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
				}
			}
			body = body[4+end+5:]
		}
	}
	d.Body = strings.TrimLeft(body, "\n")
	lines := strings.Split(body, "\n")
	ids := map[string]int{}
	var para []string
	flush := func() {
		if len(para) > 0 {
			d.Blocks = append(d.Blocks, Block{Kind: Paragraph, Text: Inlines(strings.Join(para, " "))})
			para = nil
		}
	}
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			flush()
		case fence.MatchString(t):
			flush()
			lang := fence.FindStringSubmatch(t)[1]
			var code []string
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "```"; i++ {
				code = append(code, lines[i])
			}
			d.Blocks = append(d.Blocks, Block{Kind: Code, Lang: lang, Code: strings.Join(code, "\n")})
		case heading.MatchString(t):
			flush()
			m := heading.FindStringSubmatch(t)
			id := Slug(plain(m[2]))
			if n := ids[id]; n > 0 {
				ids[id] = n + 1
				id = id + "-" + itoa(n)
			} else {
				ids[id] = 1
			}
			d.Blocks = append(d.Blocks, Block{Kind: Heading, Level: len(m[1]), ID: id, Text: Inlines(m[2])})
		case t == "---":
			flush()
			d.Blocks = append(d.Blocks, Block{Kind: Rule})
		case strings.HasPrefix(t, ">"):
			flush()
			var items [][]Inline
			var cur []string
			for ; i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">"); i++ {
				s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), ">"))
				if s == "" {
					if len(cur) > 0 {
						items = append(items, Inlines(strings.Join(cur, " ")))
						cur = nil
					}
					continue
				}
				cur = append(cur, s)
			}
			if len(cur) > 0 {
				items = append(items, Inlines(strings.Join(cur, " ")))
			}
			i--
			d.Blocks = append(d.Blocks, Block{Kind: Note, Items: items})
		case bullet.MatchString(t) || number.MatchString(t):
			flush()
			re, kind := bullet, Bullets
			if !bullet.MatchString(t) {
				re, kind = number, Numbers
			}
			var items [][]Inline
			var cur []string
			for ; i < len(lines); i++ {
				s := strings.TrimSpace(lines[i])
				if s == "" {
					break
				}
				if m := re.FindStringSubmatch(s); m != nil {
					if len(cur) > 0 {
						items = append(items, Inlines(strings.Join(cur, " ")))
					}
					cur = []string{m[1]}
					continue
				}
				if heading.MatchString(s) || fence.MatchString(s) || strings.HasPrefix(s, "|") || strings.HasPrefix(s, ">") {
					break
				}
				cur = append(cur, s) // continuation line
			}
			if len(cur) > 0 {
				items = append(items, Inlines(strings.Join(cur, " ")))
			}
			i--
			d.Blocks = append(d.Blocks, Block{Kind: kind, Items: items})
		case strings.HasPrefix(t, "|") && i+1 < len(lines) && tableSep.MatchString(strings.TrimSpace(lines[i+1])):
			flush()
			b := Block{Kind: Table, Header: cells(t)}
			for i += 2; i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|"); i++ {
				b.Rows = append(b.Rows, cells(strings.TrimSpace(lines[i])))
			}
			i--
			d.Blocks = append(d.Blocks, b)
		default:
			para = append(para, t)
		}
	}
	flush()
	return d
}

func cells(row string) [][]Inline {
	row = strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|")
	row = strings.ReplaceAll(row, `\|`, "\x00") // escaped pipe: literal, never a separator
	// split on | that is not inside `code`
	var out [][]Inline
	var cur strings.Builder
	inCode := false
	for _, r := range row {
		switch {
		case r == '`':
			inCode = !inCode
			cur.WriteRune(r)
		case r == '|' && !inCode:
			out = append(out, cell(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(out, cell(cur.String()))
}

func cell(s string) []Inline {
	return Inlines(strings.ReplaceAll(strings.TrimSpace(s), "\x00", "|"))
}

// Inlines splits text into spans.
func Inlines(s string) []Inline {
	var out []Inline
	last := 0
	for _, m := range inlineRe.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > last {
			out = append(out, Inline{Kind: Text, Text: s[last:m[0]]})
		}
		switch {
		case m[2] >= 0:
			out = append(out, Inline{Kind: Mono, Text: s[m[2]:m[3]]})
		case m[4] >= 0:
			out = append(out, Inline{Kind: Strong, Text: s[m[4]:m[5]]})
		case m[6] >= 0:
			out = append(out, Inline{Kind: Em, Text: s[m[6]:m[7]]})
		default:
			out = append(out, Inline{Kind: Link, Text: s[m[8]:m[9]], Href: s[m[10]:m[11]]})
		}
		last = m[1]
	}
	if last < len(s) {
		out = append(out, Inline{Kind: Text, Text: s[last:]})
	}
	return out
}

func plain(s string) string {
	var b strings.Builder
	for _, in := range Inlines(s) {
		b.WriteString(in.Text)
	}
	return b.String()
}

// Plain is the text of spans without markup.
func Plain(in []Inline) string {
	var b strings.Builder
	for _, x := range in {
		b.WriteString(x.Text)
	}
	return b.String()
}

// Slug makes a heading anchor: lowercase letters and digits joined by dashes (Cyrillic kept). An id must start with
// a letter, so numbered headings ("3. DNS") get the prefix "h-" ("h-3-dns").
func Slug(s string) string {
	out := strings.Trim(slugDrop.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if out != "" && out[0] >= '0' && out[0] <= '9' {
		out = "h-" + out
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}
