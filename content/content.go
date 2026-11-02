// Package content embeds the Markdown sources: content/<lang>/<page>.md. They are the single source for the HTML
// pages, the raw .md routes and llms.txt / llms-full.txt.
package content

import "embed"

//go:embed en ru
var FS embed.FS
