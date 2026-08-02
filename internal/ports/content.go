package ports

import "context"

// Content is marketing content (fixtures now → DB later).
type Content interface {
	Brand(ctx context.Context) (Brand, error)
	HomeMeta(ctx context.Context) (HomeMeta, error)
	Advantages(ctx context.Context) ([]Advantage, error)
	Solutions(ctx context.Context) ([]Solution, error)
	SolutionBySlug(ctx context.Context, slug string) (*Solution, error)
	BlogPosts(ctx context.Context) ([]BlogPost, error)
	BlogPostBySlug(ctx context.Context, slug string) (*BlogPost, error)
	About(ctx context.Context) (AboutPage, error)
	Contacts(ctx context.Context) (ContactPage, error)
	LegalDocument(ctx context.Context, id string) (*LegalDocument, error)
}

type Brand struct {
	Name        string `json:"name"`
	LegalName   string `json:"legalName"`
	Tagline     string `json:"tagline"`
	Offer       string `json:"offer"`
	Description string `json:"description"`
}

type Advantage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Glyph       string `json:"glyph"`
}

type SolutionSection struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Metric struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Note  string `json:"note"`
}

type PathStep struct {
	Year  string `json:"year"`
	Label string `json:"label"`
	Note  string `json:"note"`
}

type Quote struct {
	Text   string `json:"text"`
	Source string `json:"source"`
}

type EvidenceItem struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type Portrait struct {
	Src  string `json:"src"`
	Aria string `json:"aria"`
}

type PageSEO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

// HomeMeta is burnout landing extras (path, metrics, quote, evidence, portrait).
type HomeMeta struct {
	SEO          PageSEO        `json:"seo"`
	PathEyebrow  string         `json:"path_eyebrow"`
	PathTitle    string         `json:"path_title"`
	MetricsAria  string         `json:"metrics_aria"`
	EvidenceAria string         `json:"evidence_aria"`
	Portrait     Portrait       `json:"portrait"`
	Path         []PathStep     `json:"path"`
	Metrics      []Metric       `json:"metrics"`
	Quote        Quote          `json:"quote"`
	Evidence     []EvidenceItem `json:"evidence"`
}

// Solution is an AuthFly module (IdP, AuthKit, …).
type Solution struct {
	ID       string            `json:"id"`
	Slug     string            `json:"slug"`
	Title    string            `json:"title"`
	Eyebrow  string            `json:"eyebrow"`
	Summary  string            `json:"summary"`
	Hero     string            `json:"hero"`
	Kind     string            `json:"kind"`
	Status   string            `json:"status"`
	Accent   string            `json:"accent"`
	Outcomes []string          `json:"outcomes"`
	Sections []SolutionSection `json:"sections"`
	Metrics  []Metric          `json:"metrics"`
	Stack    []string          `json:"stack"`
}

type BlogPost struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Excerpt     string   `json:"excerpt"`
	Date        string   `json:"date"`
	ReadMinutes int      `json:"readMinutes"`
	Tags        []string `json:"tags"`
	Featured    bool     `json:"featured"`
	Body        []string `json:"body"`
}

type AboutPage struct {
	Title   string   `json:"title"`
	Lead    string   `json:"lead"`
	Body    []string `json:"body"`
	Pillars []string `json:"pillars"`
}

type ContactChannel struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Value string `json:"value"`
	Href  string `json:"href"`
	Note  string `json:"note"`
}

type ContactPage struct {
	SEO               PageSEO          `json:"seo"`
	Title             string           `json:"title"`
	Lead              string           `json:"lead"`
	ChannelsTitle     string           `json:"channels_title"`
	AvailabilityTitle string           `json:"availability_title"`
	Channels          []ContactChannel `json:"channels"`
	Availability      []string         `json:"availability"`
	Note              string           `json:"note"`
}

type LegalSection struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Points  []string `json:"points"`
}

type LegalDocument struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	ShortTitle    string         `json:"shortTitle"`
	Version       string         `json:"version"`
	EffectiveFrom string         `json:"effectiveFrom"`
	Lead          string         `json:"lead"`
	Sections      []LegalSection `json:"sections"`
}
