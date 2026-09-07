package contracts

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// The template repository's CI supplies its exact candidate checkout. Ordinary
// unit tests stay hermetic and use their own fixtures.
func TestTemplateProducerConsumer(t *testing.T) {
	root := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if root == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE to the exact template checkout")
	}
	inventory := map[string][]byte{}
	err := filepath.WalkDir(filepath.Join(root, "template"), func(p string, e fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if e.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(filepath.Join(root, "template"), p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		inventory[filepath.ToSlash(rel)] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadManifest(inventory[ManifestPath], inventory)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(manifest, inventory, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Types) != 6 || len(catalog.Bundles) != 12 {
		t.Fatalf("incomplete producer: %d types, %d bundles", len(catalog.Types), len(catalog.Bundles))
	}
	for _, preset := range []string{"core", "docs", "full", "legacy"} {
		t.Run(preset, func(t *testing.T) {
			selected, err := manifest.Select(preset, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			subset := map[string][]byte{}
			for p, entry := range manifest.Files {
				if selected.Has(entry.Component) {
					subset[p] = inventory[p]
				}
			}
			declared, err := ReadManifest(subset[ManifestPath], nil)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := LoadCatalog(declared, subset, &selected)
			if err != nil {
				t.Fatal(err)
			}
			if (len(actual.Types) > 0) != selected.Has("documents") || (len(actual.Bundles) > 0) != selected.Has("flows") {
				t.Fatal("unselected definitions were loaded")
			}
		})
	}
}
