package public

import "embed"

//go:embed **/*.css **/*.png
var Assets embed.FS
