package static

import "embed"

//go:embed index.html component.js index.js index.css
var FS embed.FS
