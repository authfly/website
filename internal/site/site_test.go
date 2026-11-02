package site

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/authfly/website/assets"
	"github.com/authfly/website/content"
	"github.com/authfly/website/internal/md"
)

func load(t *testing.T) *Content {
	t.Helper()
	c, err := Load(content.FS)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// en and ru carry the same pages and dictionary keys, and every page is in the navigation.
func TestLocalesInStep(t *testing.T) {
	c := load(t)
	if errs := c.Check(); len(errs) > 0 {
		t.Fatal(strings.Join(errs, "\n"))
	}
	for slug, en := range c.Pages["en"] {
		ru := c.Pages["ru"][slug]
		if en.Description == "" || ru.Description == "" {
			t.Errorf("%q: description missing", slug)
		}
		if en.Doc.Body == ru.Doc.Body {
			t.Errorf("%q: ru is a copy of en", slug)
		}
		if a, b := count(en.Doc, md.Code), count(ru.Doc, md.Code); a != b {
			t.Errorf("%q: code blocks en %d ru %d", slug, a, b)
		}
		if a, b := count(en.Doc, md.Table), count(ru.Doc, md.Table); a != b {
			t.Errorf("%q: tables en %d ru %d", slug, a, b)
		}
	}
}

func count(d md.Doc, kind string) int {
	n := 0
	for _, b := range d.Blocks {
		if b.Kind == kind {
			n++
		}
	}
	return n
}

func links(d md.Doc) []string {
	var out []string
	add := func(in []md.Inline) {
		for _, s := range in {
			if s.Kind == md.Link {
				out = append(out, s.Href)
			}
		}
	}
	for _, b := range d.Blocks {
		add(b.Text)
		for _, it := range b.Items {
			add(it)
		}
		for _, c := range b.Header {
			add(c)
		}
		for _, r := range b.Rows {
			for _, c := range r {
				add(c)
			}
		}
	}
	return out
}

// Every internal link resolves to a page (and its anchor), and a page links only inside its own locale.
func TestInternalLinks(t *testing.T) {
	c := load(t)
	srv := &Server{C: c, Assets: assets.FS, Origin: "http://x"}
	h := srv.Handler()
	for _, lang := range Locales {
		for slug, p := range c.Pages[lang] {
			for _, href := range links(p.Doc) {
				if !strings.HasPrefix(href, "/") {
					continue
				}
				path, frag, _ := strings.Cut(href, "#")
				if strings.HasPrefix(path, "/en/") || strings.HasPrefix(path, "/ru/") {
					if !strings.HasPrefix(path, "/"+lang+"/") {
						t.Errorf("%s/%s: link %s leaves the locale", lang, slug, href)
					}
				}
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
				if rec.Code != 200 {
					t.Errorf("%s/%s: link %s -> %d", lang, slug, href, rec.Code)
					continue
				}
				if frag != "" {
					target := c.Pages[lang][strings.TrimPrefix(path, "/"+lang+"/")]
					if target == nil || !hasID(target.Doc, frag) {
						t.Errorf("%s/%s: anchor %s missing", lang, slug, href)
					}
				}
			}
		}
	}
}

func hasID(d md.Doc, id string) bool {
	for _, b := range d.Blocks {
		if b.Kind == md.Heading && b.ID == id {
			return true
		}
	}
	return false
}

var hrefRe = regexp.MustCompile(`href="([^"]+)"`)

// Every page renders, every href in the rendered HTML (navigation, footer, content) resolves, raw .md matches.
func TestRenderedPages(t *testing.T) {
	c := load(t)
	srv := &Server{C: c, Assets: assets.FS, Origin: "http://x"}
	h := srv.Handler()
	get := func(path string) (int, string, http.Header) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		b, _ := io.ReadAll(rec.Result().Body)
		return rec.Code, string(b), rec.Header()
	}
	seen := map[string]bool{}
	for _, lang := range Locales {
		for _, p := range c.Ordered(lang) {
			code, body, _ := get(URL(lang, p.Slug))
			if code != 200 {
				t.Fatalf("%s: %d", URL(lang, p.Slug), code)
			}
			if strings.Contains(body, "<script>") || strings.Contains(body, " style=") {
				t.Errorf("%s: inline script or style", URL(lang, p.Slug))
			}
			for _, m := range hrefRe.FindAllStringSubmatch(body, -1) {
				u := strings.SplitN(m[1], "#", 2)[0]
				u = strings.SplitN(u, "?", 2)[0]
				if u == "" || strings.HasPrefix(u, "http") || seen[u] {
					continue
				}
				seen[u] = true
				if code, _, _ := get(u); code != 200 {
					t.Errorf("%s: href %s -> %d", URL(lang, p.Slug), m[1], code)
				}
			}
			code, raw, hd := get(MDURL(lang, p.Slug))
			if code != 200 || raw != p.Doc.Source || !strings.HasPrefix(hd.Get("Content-Type"), "text/markdown") {
				t.Errorf("%s: raw markdown mismatch", MDURL(lang, p.Slug))
			}
		}
	}
	for _, path := range []string{"/llms.txt", "/ru/llms.txt", "/llms-full.txt", "/ru/llms-full.txt", "/sitemap.xml", "/robots.txt"} {
		if code, body, _ := get(path); code != 200 || len(body) < 40 {
			t.Errorf("%s: %d", path, code)
		}
	}
	if code, _, _ := get("/en/does-not-exist"); code != 404 {
		t.Errorf("404 page: %d", code)
	}
}

// No real IP addresses (documentation ranges and loopback only), no stale license or product claims.
func TestContentHygiene(t *testing.T) {
	c := load(t)
	banned := regexp.MustCompile(`(?i)\bMIT\b|authkit-ts|\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	allowedIP := map[string]bool{"203.0.113.10": true, "127.0.0.1": true, "10.118.42.1": true, "0.0.0.0": true}
	for _, lang := range Locales {
		for slug, p := range c.Pages[lang] {
			for _, m := range banned.FindAllString(p.Doc.Source, -1) {
				if allowedIP[m] || (strings.EqualFold(m, "authkit-ts") && slug == "legal") {
					continue
				}
				t.Errorf("%s/%s: %q", lang, slug, m)
			}
		}
	}
}
