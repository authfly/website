package site

import (
	"bytes"
	"io/fs"
	"net/http"
	"strings"

	"github.com/a-h/templ"

	"github.com/authfly/website/internal/md"
	"github.com/authfly/website/internal/views"
)

// SourceBase is where "edit this page" points.
const SourceBase = "https://github.com/authfly/website/blob/main/content/"

// Server serves the docs.
type Server struct {
	C      *Content
	Assets fs.FS
	Origin string // absolute origin for canonical links and llms.txt, e.g. https://authfly.vercel.app
}

// legacy paths of the previous marketing site -> new pages
var legacy = map[string]string{
	"/projects": "targets", "/solutions": "targets", "/about": "concepts", "/advantages": "concepts",
	"/blog": "releases", "/contacts": "legal", "/legal": "legal", "/legal/privacy": "legal", "/legal/terms": "legal",
	"/privacy": "legal", "/terms": "legal", "/docs": "quickstart",
}

// Handler returns the site mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler(s.Assets)))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/assets/favicon.svg", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("User-agent: *\nAllow: /\nSitemap: " + s.Origin + "/sitemap.xml\n"))
	})
	mux.HandleFunc("GET /sitemap.xml", s.sitemap)
	mux.HandleFunc("GET /llms.txt", func(w http.ResponseWriter, r *http.Request) { s.llms(w, "en") })
	mux.HandleFunc("GET /ru/llms.txt", func(w http.ResponseWriter, r *http.Request) { s.llms(w, "ru") })
	mux.HandleFunc("GET /llms-full.txt", func(w http.ResponseWriter, r *http.Request) { s.llmsFull(w, "en") })
	mux.HandleFunc("GET /ru/llms-full.txt", func(w http.ResponseWriter, r *http.Request) { s.llmsFull(w, "ru") })
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("GET /", s.page)
	return secure(mux)
}

func secure(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		hd.Set("X-Frame-Options", "DENY")
		h.ServeHTTP(w, r)
	})
}

func assetHandler(a fs.FS) http.Handler {
	fsh := http.FileServerFS(a)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fsh.ServeHTTP(w, r)
	})
}

func preferred(r *http.Request) string {
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Accept-Language")), "ru") {
		return "ru"
	}
	return "en"
}

func (s *Server) page(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p == "/" {
		http.Redirect(w, r, "/"+preferred(r)+"/", http.StatusFound)
		return
	}
	if to, ok := legacy[strings.TrimSuffix(p, "/")]; ok {
		http.Redirect(w, r, URL(preferred(r), to), http.StatusMovedPermanently)
		return
	}
	parts := strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)
	lang := parts[0]
	if _, ok := s.C.Pages[lang]; !ok {
		s.notFound(w, r, preferred(r))
		return
	}
	if len(parts) == 1 {
		http.Redirect(w, r, "/"+lang+"/", http.StatusMovedPermanently)
		return
	}
	slug := parts[1]
	if strings.HasSuffix(slug, ".md") {
		slug = strings.TrimSuffix(slug, ".md")
		if slug == "index" {
			slug = ""
		}
		if pg := s.C.Pages[lang][slug]; pg != nil {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Cache-Control", "public, max-age=300")
			w.Write([]byte(pg.Doc.Source))
			return
		}
		s.notFound(w, r, lang)
		return
	}
	if slug != "" && strings.HasSuffix(slug, "/") {
		http.Redirect(w, r, "/"+lang+"/"+strings.TrimSuffix(slug, "/"), http.StatusMovedPermanently)
		return
	}
	pg := s.C.Pages[lang][slug]
	if pg == nil {
		s.notFound(w, r, lang)
		return
	}
	sh := s.shell(lang, slug, pg.Title, pg.Description, pg.Doc)
	w.Header().Set("Cache-Control", "public, max-age=300")
	var c templ.Component = views.DocPage(sh, pg.Doc.Blocks)
	if slug == "" {
		c = views.HomePage(sh, pg.Doc.Blocks)
	}
	render(w, r, http.StatusOK, c)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request, lang string) {
	t := s.C.Dict[lang]
	sh := s.shell(lang, "", t["notfound.title"], "", md.Doc{})
	sh.Toc, sh.Prev, sh.Next = nil, nil, nil
	render(w, r, http.StatusNotFound, views.NotFound(sh))
}

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	var buf bytes.Buffer
	if err := c.Render(r.Context(), &buf); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func other(lang string) string {
	if lang == "en" {
		return "ru"
	}
	return "en"
}

func (s *Server) shell(lang, slug, title, desc string, doc md.Doc) views.Shell {
	t := s.C.Dict[lang]
	full := title + " | " + t["site.name"]
	if slug == "" {
		full = t["site.name"] + " | " + t["site.tagline"]
	}
	if desc == "" {
		desc = t["site.description"]
	}
	file := slug
	if file == "" {
		file = "index"
	}
	sh := views.Shell{
		Lang: lang, Title: full, Description: desc, Canonical: s.Origin + URL(lang, slug), Path: URL(lang, slug),
		AltLang: other(lang), AltHref: URL(other(lang), slug), MDHref: MDURL(lang, slug),
		SourceHref: SourceBase + lang + "/" + file + ".md", T: t, Home: slug == "",
	}
	ordered := s.C.Ordered(lang)
	for _, sec := range Nav {
		g := views.Group{Label: t["nav."+sec.Key]}
		for _, sl := range sec.Slugs {
			if p := s.C.Pages[lang][sl]; p != nil {
				g.Links = append(g.Links, views.Link{Href: URL(lang, sl), Title: navTitle(p), Current: sl == slug})
			}
		}
		sh.Nav = append(sh.Nav, g)
	}
	for i, p := range ordered {
		if p.Slug != slug {
			continue
		}
		if i > 0 {
			sh.Prev = &views.Link{Href: URL(lang, ordered[i-1].Slug), Title: navTitle(ordered[i-1])}
		}
		if i+1 < len(ordered) {
			sh.Next = &views.Link{Href: URL(lang, ordered[i+1].Slug), Title: navTitle(ordered[i+1])}
		}
	}
	for _, b := range doc.Blocks {
		if b.Kind == md.Heading && (b.Level == 2 || b.Level == 3) {
			sh.Toc = append(sh.Toc, views.TocItem{ID: b.ID, Title: md.Plain(b.Text), Sub: b.Level == 3})
		}
	}
	return sh
}

func navTitle(p *Page) string {
	if v := p.Doc.Meta["nav"]; v != "" {
		return v
	}
	return p.Title
}

func (s *Server) llms(w http.ResponseWriter, lang string) {
	t := s.C.Dict[lang]
	var b strings.Builder
	b.WriteString("# " + t["site.name"] + "\n\n> " + t["llms.summary"] + "\n\n" + t["llms.intro"] + "\n")
	for _, sec := range Nav {
		b.WriteString("\n## " + t["nav."+sec.Key] + "\n\n")
		for _, sl := range sec.Slugs {
			if p := s.C.Pages[lang][sl]; p != nil {
				b.WriteString("- [" + p.Title + "](" + s.Origin + MDURL(lang, sl) + "): " + p.Description + "\n")
			}
		}
	}
	b.WriteString("\n## " + t["llms.optional"] + "\n\n")
	b.WriteString("- [llms-full.txt](" + s.Origin + strings.TrimPrefix("/"+lang, "/en") + "/llms-full.txt): " + t["llms.full"] + "\n")
	b.WriteString("- [" + t["llms.other"] + "](" + s.Origin + strings.TrimPrefix("/"+other(lang), "/en") + "/llms.txt)\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(b.String()))
}

func (s *Server) llmsFull(w http.ResponseWriter, lang string) {
	var b strings.Builder
	for i, p := range s.C.Ordered(lang) {
		if i > 0 {
			b.WriteString("\n\n---\n\n")
		}
		b.WriteString("<!-- source: " + s.Origin + MDURL(lang, p.Slug) + " -->\n\n")
		b.WriteString(strings.TrimSpace(p.Doc.Body) + "\n")
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(b.String()))
}

func (s *Server) sitemap(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, l := range Locales {
		for _, p := range s.C.Ordered(l) {
			b.WriteString("<url><loc>" + s.Origin + URL(l, p.Slug) + "</loc></url>\n")
		}
	}
	b.WriteString("</urlset>\n")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(b.String()))
}
