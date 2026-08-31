package dir

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

func TestAddParser(t *testing.T) {
	oldParsers := knownParser
	defer func() { knownParser = oldParsers }()

	initialCount := len(knownParser)
	AddParser(func(filePath, basePath string, data *parser.DomainMap) error {
		return nil
	})
	if len(knownParser) != initialCount+1 {
		t.Error("AddParser failed to add a parser")
	}
}

func TestParseDir(t *testing.T) {
	oldParsers := knownParser
	defer func() { knownParser = oldParsers }()
	knownParser = nil

	dm := &parser.DomainMap{}
	called := false
	AddParser(func(dirPath, basePath string, data *parser.DomainMap) error {
		called = true
		return nil
	})

	err := ParseDir(".", ".", dm)
	if err != nil {
		t.Errorf("ParseDir failed: %v", err)
	}
	if !called {
		t.Error("ParseDir did not call the added parser")
	}
}

func TestParseDirRec(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gotext-test-rec-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	subDir := filepath.Join(tmpDir, "sub")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	excludeDir := filepath.Join(tmpDir, "exclude")
	err = os.Mkdir(excludeDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	excludeChild := filepath.Join(excludeDir, "child")
	err = os.Mkdir(excludeChild, 0755)
	if err != nil {
		t.Fatal(err)
	}

	excludeSibling := filepath.Join(tmpDir, "excluded")
	err = os.Mkdir(excludeSibling, 0755)
	if err != nil {
		t.Fatal(err)
	}

	oldParsers := knownParser
	defer func() { knownParser = oldParsers }()

	dm := &parser.DomainMap{}
	visited := make(map[string]bool)
	knownParser = []ParseDirFunc{func(dirPath, basePath string, data *parser.DomainMap) error {
		visited[dirPath] = true
		return nil
	}}

	err = ParseDirRec(tmpDir, []string{"", ".", "." + string(filepath.Separator) + "exclude"}, dm, true)
	if err != nil {
		t.Errorf("ParseDirRec failed: %v", err)
	}

	for _, path := range []string{tmpDir, subDir, excludeSibling} {
		if !visited[path] {
			t.Errorf("Expected parser call for %s", path)
		}
	}
	for _, path := range []string{excludeDir, excludeChild} {
		if visited[path] {
			t.Errorf("Unexpected parser call for excluded path %s", path)
		}
	}
	if len(visited) != 3 {
		t.Errorf("Expected 3 parser calls, got %d", len(visited))
	}
}

func TestParseDirRecExclusionSpellings(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	targetChild := filepath.Join(target, "child")
	sibling := filepath.Join(root, "targeted")
	for _, path := range []string{targetChild, sibling} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	exclusions := []string{
		"target",
		filepath.Join(".", "target"),
		filepath.Join("keep", "..", "target"),
		target,
		filepath.Join(root, "keep", "..", "target"),
	}
	for _, exclusion := range exclusions {
		t.Run(exclusion, func(t *testing.T) {
			oldParsers := knownParser
			t.Cleanup(func() { knownParser = oldParsers })

			visited := make(map[string]bool)
			knownParser = []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
				visited[path] = true
				return nil
			}}

			if err := ParseDirRec(root, []string{exclusion}, &parser.DomainMap{}, false); err != nil {
				t.Fatal(err)
			}
			if visited[target] || visited[targetChild] {
				t.Fatalf("exclusion %q visited target directory", exclusion)
			}
			if !visited[root] || !visited[sibling] {
				t.Fatalf("exclusion %q pruned root or prefix sibling", exclusion)
			}
		})
	}
}

func TestParseDirRecCallbackOrderingAndError(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.MkdirAll(filepath.Join(child, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sibling"), 0755); err != nil {
		t.Fatal(err)
	}

	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	stop := errors.New("stop traversal")
	events := make([]string, 0, 4)
	knownParser = []ParseDirFunc{
		func(path, _ string, _ *parser.DomainMap) error {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			events = append(events, "first:"+rel)
			return nil
		},
		func(path, _ string, _ *parser.DomainMap) error {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			events = append(events, "second:"+rel)
			if path == child {
				return stop
			}
			return nil
		},
		func(path, _ string, _ *parser.DomainMap) error {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			events = append(events, "third:"+rel)
			return nil
		},
	}

	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err != stop {
		t.Fatalf("ParseDirRec error = %v, want %v", err, stop)
	}
	want := []string{"first:.", "second:.", "third:.", "first:child", "second:child"}
	if len(events) != len(want) {
		t.Fatalf("callback events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("callback events = %v, want %v", events, want)
		}
	}
}

func TestParseDirRecRootCallbackError(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "child"), 0755); err != nil {
		t.Fatal(err)
	}

	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	stop := errors.New("root callback failed")
	calls := 0
	knownParser = []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
		calls++
		if path == root {
			return stop
		}
		return nil
	}}

	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err != stop {
		t.Fatalf("ParseDirRec error = %v, want %v", err, stop)
	}
	if calls != 1 {
		t.Fatalf("callback calls after root failure = %d, want 1", calls)
	}
}

func TestParseDirRecMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	called := false
	knownParser = []ParseDirFunc{func(string, string, *parser.DomainMap) error {
		called = true
		return nil
	}}

	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err == nil {
		t.Fatal("ParseDirRec unexpectedly succeeded for missing root")
	}
	if called {
		t.Fatal("ParseDirRec invoked a callback for a missing root")
	}
}

func TestParseDirRecDoesNotFollowSymlink(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(filepath.Join(root, "inside"), 0755); err != nil {
		t.Fatal(err)
	}
	outsideChild := filepath.Join(outside, "child")
	if err := os.MkdirAll(outsideChild, 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "outside-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	visited := make(map[string]bool)
	knownParser = []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
		visited[path] = true
		return nil
	}}
	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err != nil {
		t.Fatal(err)
	}
	if visited[link] || visited[outside] || visited[outsideChild] {
		t.Fatalf("symlink traversal escaped root: %v", visited)
	}
	for path := range visited {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("callback outside root: %s", path)
		}
	}
}

func FuzzParseDirRecExcludeNormalization(f *testing.F) {
	root := f.TempDir()
	target := filepath.Join(root, "target")
	targeted := filepath.Join(root, "targeted")
	for _, path := range []string{filepath.Join(target, "child"), filepath.Join(targeted, "child")} {
		if err := os.MkdirAll(path, 0755); err != nil {
			f.Fatal(err)
		}
	}

	f.Add("target")
	f.Add(filepath.Join(".", "target"))
	f.Add(filepath.Join("keep", "..", "target"))
	f.Add(target)
	f.Add(filepath.Join(root, "keep", "..", "target"))
	f.Add(filepath.Join("..", filepath.Base(root), "target"))
	f.Add("")
	f.Add("\x00")

	oldParsers := knownParser
	f.Cleanup(func() { knownParser = oldParsers })

	f.Fuzz(func(t *testing.T, spelling string) {
		if len(spelling) > 64*1024 {
			return
		}
		visited := make(map[string]bool)
		knownParser = []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
			visited[path] = true
			return nil
		}}

		if err := ParseDirRec(root, []string{spelling}, &parser.DomainMap{}, false); err != nil {
			t.Fatalf("ParseDirRec(%q): %v", spelling, err)
		}
		for path := range visited {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				t.Fatalf("callback outside root for exclusion %q: %s", spelling, path)
			}
		}

		normalized := spelling
		if !filepath.IsAbs(normalized) {
			normalized = filepath.Join(root, normalized)
		}
		if filepath.Clean(normalized) == target {
			if visited[target] || !visited[targeted] {
				t.Fatalf("normalized exclusion %q did not prune only target: %v", spelling, visited)
			}
		}
	})
}

func TestGoFile_InspectFile_Coverage(t *testing.T) {
	// Minimal coverage for the InspectFile switch
	g := &GoFile{}
	g.InspectFile(nil) // Default case
}
