// Package assets embeds the built stylesheet (bun run build:css), theme.js and the favicon.
package assets

import "embed"

//go:embed app.css theme.js favicon.svg
var FS embed.FS
