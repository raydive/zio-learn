// Package assets embeds the resources shipped with the playground binary.
package assets

import "embed"

//go:embed web config.json examples
var files embed.FS

// Files returns the application's bundled configuration, examples and web UI.
func Files() embed.FS {
	return files
}
