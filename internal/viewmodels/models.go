package viewmodels

import (
	"fmt"
	"strconv"
	"time"

	"github.com/authfly/website/internal/fixtures"
	"github.com/authfly/website/internal/ports"
)

// LayoutModel is shared chrome for every full page.
type LayoutModel struct {
	Title       string
	Description string
	OgImage     string
	Canonical   string
	Brand       ports.Brand
	Offer       string
	Path        string
	Copy        fixtures.Locale
	Lang        string
}

// HomeModel is the landing page.
type HomeModel struct {
	Layout     LayoutModel
	Meta       ports.HomeMeta
	Advantages []ports.Advantage
	Solutions  []SolutionCardModel
	Posts      []BlogCardModel
}

// AboutModel is the about page.
type AboutModel struct {
	Layout LayoutModel
	Page   ports.AboutPage
	Brand  ports.Brand
}

// ContactsModel is the contacts page (info only — no forms).
type ContactsModel struct {
	Layout LayoutModel
	Page   ports.ContactPage
}

// LegalModel is a legal document page.
type LegalModel struct {
	Layout LayoutModel
	Doc    ports.LegalDocument
}

// NotFoundModel is the 404 page.
type NotFoundModel struct {
	Layout LayoutModel
	Path   string
}

// SolutionCardModel is a compact module tile.
type SolutionCardModel struct {
	Slug    string
	Title   string
	Eyebrow string
	Summary string
	Status  string
	Accent  string
	Href    string
}

// SolutionModel is a module detail page.
type SolutionModel struct {
	Layout   LayoutModel
	Solution ports.Solution
}

// SolutionsModel is the modules index.
type SolutionsModel struct {
	Layout    LayoutModel
	Solutions []SolutionCardModel
}

// BlogCardModel is a blog list/card entry.
type BlogCardModel struct {
	Slug        string
	Title       string
	Excerpt     string
	Date        string
	ReadMinutes int
	Tags        []string
	Href        string
	Featured    bool
	Lang        string
}

// BlogModel is the blog index.
type BlogModel struct {
	Layout LayoutModel
	Posts  []BlogCardModel
}

// BlogPostModel is a single article page.
type BlogPostModel struct {
	Layout LayoutModel
	Post   ports.BlogPost
}

// WithBrand builds "Page — Brand" titles.
func WithBrand(title, brand string) string {
	if brand == "" {
		return title
	}
	if title == "" {
		return brand
	}
	return fmt.Sprintf("%s — %s", title, brand)
}

// Itoa formats an int for templ text nodes.
func Itoa(n int) string {
	return strconv.Itoa(n)
}

var monthsRU = [...]string{
	"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря",
}

var monthsEN = [...]string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// FormatDate renders an ISO date. lang "en" uses English month names.
func FormatDate(iso string, lang ...string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	code := fixtures.ActiveLocale()
	if len(lang) > 0 && lang[0] != "" {
		code = lang[0]
	}
	if code == "en" {
		return fmt.Sprintf("%d %s %d", t.Day(), monthsEN[int(t.Month())-1], t.Year())
	}
	return fmt.Sprintf("%d %s %d", t.Day(), monthsRU[int(t.Month())-1], t.Year())
}

// MinutesWord returns the locale "min" label for cards without a full layout.
func MinutesWord() string {
	return fixtures.MustLocale(fixtures.ActiveLocale()).Common.Minutes
}

// MoreWord returns the locale "details" label for cards.
func MoreWord() string {
	return fixtures.MustLocale(fixtures.ActiveLocale()).CommonMore
}
