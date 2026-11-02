// Package site loads the embedded Markdown content and the UI dictionaries and serves the docs site.
package site

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/authfly/website/internal/md"
)

// Locales served, default first.
var Locales = []string{"en", "ru"}

// Section groups pages in the sidebar; Key names the dictionary label "nav.<key>".
type Section struct {
	Key   string
	Slugs []string
}

// Nav is the information architecture: one list for every locale, so en and ru stay in step (a test enforces it).
var Nav = []Section{
	{Key: "start", Slugs: []string{"", "quickstart", "concepts"}},
	{Key: "guides", Slugs: []string{"guides/install-authfly-stalwart", "guides/react-docker"}},
	{Key: "reference", Slugs: []string{"targets", "reference/authfly-yaml", "reference/cli", "reference/http-api", "reference/sdks"}},
	{Key: "trust", Slugs: []string{"security", "policy-packs"}},
	{Key: "project", Slugs: []string{"releases", "legal"}},
}

// Page is one parsed content file.
type Page struct {
	Slug, Lang, Title, Description string
	Doc                            md.Doc
}

// Content holds every page and dictionary.
type Content struct {
	Pages map[string]map[string]*Page // lang -> slug -> page
	Dict  map[string]map[string]string
}

//go:embed i18n/*.json
var dictFS embed.FS

// Load reads content/<lang>/**/*.md ("index.md" is the home page).
func Load(content fs.FS) (*Content, error) {
	c := &Content{Pages: map[string]map[string]*Page{}, Dict: map[string]map[string]string{}}
	for _, lang := range Locales {
		raw, err := dictFS.ReadFile("i18n/" + lang + ".json")
		if err != nil {
			return nil, err
		}
		d := map[string]string{}
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, fmt.Errorf("i18n/%s.json: %w", lang, err)
		}
		c.Dict[lang] = d
		c.Pages[lang] = map[string]*Page{}
		err = fs.WalkDir(content, lang, func(p string, e fs.DirEntry, err error) error {
			if err != nil || e.IsDir() || path.Ext(p) != ".md" {
				return err
			}
			b, err := fs.ReadFile(content, p)
			if err != nil {
				return err
			}
			slug := strings.TrimSuffix(strings.TrimPrefix(p, lang+"/"), ".md")
			if slug == "index" {
				slug = ""
			}
			doc := md.Parse(string(b))
			if doc.Meta["title"] == "" {
				return fmt.Errorf("%s: front matter title missing", p)
			}
			c.Pages[lang][slug] = &Page{Slug: slug, Lang: lang, Title: doc.Meta["title"], Description: doc.Meta["description"], Doc: doc}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Ordered returns the pages of a locale in navigation order.
func (c *Content) Ordered(lang string) []*Page {
	var out []*Page
	for _, s := range Nav {
		for _, slug := range s.Slugs {
			if p := c.Pages[lang][slug]; p != nil {
				out = append(out, p)
			}
		}
	}
	return out
}

// Check reports pages missing from a locale, pages outside the navigation and dictionary keys that differ.
func (c *Content) Check() []string {
	var errs []string
	inNav := map[string]bool{}
	for _, s := range Nav {
		for _, slug := range s.Slugs {
			inNav[slug] = true
			for _, l := range Locales {
				if c.Pages[l][slug] == nil {
					errs = append(errs, fmt.Sprintf("%s: page %q missing", l, slug))
				}
			}
		}
	}
	for _, l := range Locales {
		for slug := range c.Pages[l] {
			if !inNav[slug] {
				errs = append(errs, fmt.Sprintf("%s: page %q not in Nav", l, slug))
			}
		}
		for k := range c.Dict[Locales[0]] {
			if _, ok := c.Dict[l][k]; !ok {
				errs = append(errs, fmt.Sprintf("%s: dictionary key %q missing", l, k))
			}
		}
	}
	sort.Strings(errs)
	return errs
}

// URL is the HTML path of a page.
func URL(lang, slug string) string { return "/" + lang + "/" + slug }

// MDURL is the raw Markdown path of a page.
func MDURL(lang, slug string) string {
	if slug == "" {
		slug = "index"
	}
	return "/" + lang + "/" + slug + ".md"
}
