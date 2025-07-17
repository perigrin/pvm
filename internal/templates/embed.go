// ABOUTME: Embedded template assets for PVM workspace initialization
// ABOUTME: Provides embed.FS access to template files at compile time

package templates

import "embed"

//go:embed assets/templates
var FS embed.FS
