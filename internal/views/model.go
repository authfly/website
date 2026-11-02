// Package views renders the docs site with UI8Kit 3 templ primitives and the Authfly templ layer only.
package views

// Version busts asset caches; set by main from the build.
var Version = "dev"

// Link is a navigation entry.
type Link struct {
	Href, Title string
	Current     bool
}

// Group is a sidebar section.
type Group struct {
	Label string
	Links []Link
}

// TocItem is one in-page anchor.
type TocItem struct {
	ID, Title string
	Sub       bool
}

// Shell is everything the layout needs around a page body.
type Shell struct {
	Lang, Title, Description, Canonical, Path string
	AltLang, AltHref                          string // the same page in the other locale
	MDHref, SourceHref                        string
	Nav                                       []Group
	Toc                                       []TocItem
	Prev, Next                                *Link
	T                                         map[string]string
	Home                                      bool
}

func (s Shell) t(k string) string {
	if v, ok := s.T[k]; ok {
		return v
	}
	return k
}
