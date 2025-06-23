package static

import "embed"

//go:embed index.html component.js index.js index.css utils.js
var FS embed.FS
