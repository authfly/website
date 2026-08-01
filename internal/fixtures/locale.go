package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	_ "embed"
)

//go:embed locale/ru.json
var ruJSON []byte

//go:embed locale/en.json
var enJSON []byte

// Locale is UI chrome copy (RU for Docker, EN for Vercel).
type Locale struct {
	BrandHomeAria string `json:"brand_home_aria"`
	SkipToContent string `json:"skip_to_content"`
	Theme         struct {
		Label         string `json:"label"`
		SwitchToDark  string `json:"switch_to_dark"`
		SwitchToLight string `json:"switch_to_light"`
	} `json:"theme"`
	Nav struct {
		Solutions     string `json:"solutions"`
		Blog          string `json:"blog"`
		About         string `json:"about"`
		Contacts      string `json:"contacts"`
		PrimaryAria   string `json:"primary_aria"`
		MobileAria    string `json:"mobile_aria"`
		MenuAria      string `json:"menu_aria"`
		CloseMenuAria string `json:"close_menu_aria"`
	} `json:"nav"`
	Footer struct {
		Sections  string `json:"sections"`
		Solutions string `json:"solutions"`
		Blog      string `json:"blog"`
		About     string `json:"about"`
		Docs      string `json:"docs"`
		Privacy   string `json:"privacy"`
		Terms     string `json:"terms"`
		Contacts  string `json:"contacts"`
	} `json:"footer"`
	Common struct {
		Back     string `json:"back"`
		ReadMore string `json:"read_more"`
		Minutes  string `json:"minutes"`
	} `json:"common"`
	Pages struct {
		Projects string `json:"projects"`
		Blog     string `json:"blog"`
		About    string `json:"about"`
		Contacts string `json:"contacts"`
		NotFound string `json:"not_found"`
	} `json:"pages"`
	Home struct {
		CTAPrimary      string `json:"cta_primary"`
		CTASecondary    string `json:"cta_secondary"`
		FeaturedLabel   string `json:"featured_label"`
		FeaturedLink    string `json:"featured_link"`
		AdvantagesTitle string `json:"advantages_title"`
		ProjectsRail    string `json:"projects_rail"`
		ProjectsAll     string `json:"projects_all"`
		BlogRail        string `json:"blog_rail"`
		BlogAll         string `json:"blog_all"`
	} `json:"home"`
	ProjectsIndex struct {
		Title string `json:"title"`
		Lead  string `json:"lead"`
	} `json:"projects_index"`
	BlogIndex struct {
		Title string `json:"title"`
		Lead  string `json:"lead"`
	} `json:"blog_index"`
	CommonMore   string `json:"common_more"`
	BackProjects string `json:"back_projects"`
	OutcomesTitle string `json:"outcomes_title"`
	MetricsTitle  string `json:"metrics_title"`
	StackTitle    string `json:"stack_title"`
}

// ActiveLocale returns APP_DEFAULT_LOCALE or "ru".
func ActiveLocale() string {
	return envOr("APP_DEFAULT_LOCALE", "ru")
}

// DefaultLocale is kept for callers; prefer ActiveLocale() when env may change.
var DefaultLocale = ActiveLocale()

var (
	localeMu    sync.RWMutex
	localeCache = map[string]Locale{}
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func LoadLocale(code string) (Locale, error) {
	if code == "" {
		code = DefaultLocale
	}
	localeMu.RLock()
	if loc, ok := localeCache[code]; ok {
		localeMu.RUnlock()
		return loc, nil
	}
	localeMu.RUnlock()

	var raw []byte
	switch code {
	case "en":
		raw = enJSON
	case "ru":
		raw = ruJSON
	default:
		raw = ruJSON
		code = "ru"
	}
	var loc Locale
	if err := json.Unmarshal(raw, &loc); err != nil {
		return Locale{}, fmt.Errorf("locale %s: %w", code, err)
	}
	localeMu.Lock()
	localeCache[code] = loc
	localeMu.Unlock()
	return loc, nil
}

func MustLocale(code string) Locale {
	loc, err := LoadLocale(code)
	if err != nil {
		panic(err)
	}
	return loc
}
