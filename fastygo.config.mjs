/**
 * Tooling config (Vite-like single file for paths & policy).
 *
 * Not a bundler / HMR runtime. Go owns the server.
 * Bun scripts (CSS, retag) read this file so paths and the retag allowlist
 * stay in one place — same shape across FastyGo apps.
 */
export default {
  server: {
    host: "127.0.0.1",
    port: 3011,
    staticDir: "web/static",
    env: {
      APP_BIND: "127.0.0.1:3011",
      APP_STATIC_DIR: "web/static",
      FIXTURE_DIR: "internal/adapters/fixture/data",
    },
  },

  templ: {
    generate: "./...",
  },

  css: {
    input: "web/static/css/input.css",
    output: "web/static/css/app.css",
    sources: [
      "internal/views/**/*.templ",
      "internal/ui/**/*.templ",
      "internal/kit/ui/**/*.templ",
      "internal/kit/ui/**/*.variants.json",
    ],
  },

  /**
   * UI registry map — where to find layers (Cursor rule: ui-registry.mdc).
   * Shapes future shadcn-like `fastygo add` / FastyGoUI extraction.
   */
  registry: {
    primitives: "internal/kit/ui",
    utils: "internal/kit/utils",
    layout: "internal/ui/layout",
    components: "internal/ui/components",
    blocks: "internal/ui/blocks",
    widgets: "internal/ui/widgets",
    variants: "internal/ui/variants",
    views: "internal/views",
    locale: "internal/fixtures/locale",
    future: {
      primitives: "github.com/fastygo/templ",
      blocks: "github.com/fastygo/blocks",
      widgets: "github.com/fastygo/widgets",
      components: "cli-copy",
      layout: "app-owned",
    },
  },

  /**
   * Retag policy — source of truth for scripts/retag-check.sh
   * and .project/retag.md.
   */
  retag: {
    skip: ["internal/kit/ui"],
    layout: ["internal/ui/layout"],
    layoutTags: [
      "html",
      "head",
      "body",
      "meta",
      "title",
      "link",
      "script",
      "noscript",
      "main",
    ],
    /**
     * Explicit registry: raw <input type="hidden"> only in listed files.
     * Adding a new hidden field requires an entry here (comment alone is not enough).
     */
    allow: [],
    baseline: ".project/retag-baseline.txt",
  },
};
