// Package atlantis embeds the portable brain documents shipped with the CLI.
package atlantis

import (
	"embed"
	"io/fs"
)

//go:embed brain
var brainSeed embed.FS

// BrainSeed returns the portable brain documents rooted at the vault layout.
func BrainSeed() fs.FS {
	seed, err := fs.Sub(brainSeed, "brain")
	if err != nil {
		panic(err)
	}
	return seed
}
