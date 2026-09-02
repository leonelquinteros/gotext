package dir

import (
	"errors"
	"os"
	"path/filepath"
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
	root := t.TempDir()
	dm := &parser.DomainMap{}

	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	var gotDir, gotBase string
	var gotData *parser.DomainMap
	knownParser = []ParseDirFunc{func(dirPath, basePath string, data *parser.DomainMap) error {
		gotDir, gotBase, gotData = dirPath, basePath, data
		return nil
	}}

	if err := ParseDir(root, root, dm); err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}
	if gotDir != root || gotBase != root || gotData != dm {
		t.Fatalf("ParseDir callback = (%q, %q, %p), want (%q, %q, %p)", gotDir, gotBase, gotData, root, root, dm)
	}
}
func TestParseDirRec(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	excludeDir := filepath.Join(tmpDir, "exclude")
	excludeChild := filepath.Join(excludeDir, "child")
	if err := os.MkdirAll(excludeChild, 0755); err != nil {
		t.Fatal(err)
	}

	excludeSibling := filepath.Join(tmpDir, "excluded")
	if err := os.Mkdir(excludeSibling, 0755); err != nil {
		t.Fatal(err)
	}

	oldParsers := knownParser
	t.Cleanup(func() { knownParser = oldParsers })

	dm := &parser.DomainMap{}
	visited := make(map[string]bool)
	knownParser = []ParseDirFunc{func(dirPath, basePath string, data *parser.DomainMap) error {
		visited[dirPath] = true
		return nil
	}}

	excludeSpelling := "." + string(filepath.Separator) + "exclude"
	if err := ParseDirRec(tmpDir, []string{"", ".", excludeSpelling}, dm, false); err != nil {
		t.Fatalf("ParseDirRec failed: %v", err)
	}

	assertVisitedDirs(t, visited, []string{tmpDir, subDir, excludeSibling})
}

func assertVisitedDirs(t *testing.T, visited map[string]bool, want []string) {
	t.Helper()
	expected := make(map[string]bool, len(want))
	for _, path := range want {
		expected[path] = true
	}
	if len(visited) != len(expected) {
		t.Fatalf("visited directories = %v, want %v", visited, expected)
	}
	for path := range expected {
		if !visited[path] {
			t.Errorf("expected parser call for %s", path)
		}
	}
	for path := range visited {
		if !expected[path] {
			t.Errorf("unexpected parser call for %s", path)
		}
	}
}

func TestParseDirRecExclusionSpellings(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	targetChild := filepath.Join(target, "child")
	prefixSibling := filepath.Join(root, "targeted")
	prefixSiblingChild := filepath.Join(prefixSibling, "child")
	for _, path := range []string{targetChild, prefixSiblingChild} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	separator := string(filepath.Separator)
	allDirs := []string{root, target, targetChild, prefixSibling, prefixSiblingChild}
	targetExcluded := []string{root, prefixSibling, prefixSiblingChild}
	cases := []struct {
		name     string
		spelling string
		want     []string
	}{
		{name: "relative", spelling: "target", want: targetExcluded},
		{name: "relative-dot", spelling: "." + separator + "target", want: targetExcluded},
		{name: "relative-parent", spelling: "keep" + separator + ".." + separator + "target", want: targetExcluded},
		{name: "absolute", spelling: target, want: targetExcluded},
		{name: "absolute-parent", spelling: filepath.Join(root, "keep") + separator + ".." + separator + "target", want: targetExcluded},
		{name: "relative-parent-root", spelling: ".." + separator + filepath.Base(root) + separator + "target", want: targetExcluded},
		{name: "root-relative", spelling: ".", want: allDirs},
		{name: "root-absolute", spelling: root, want: allDirs},
		{name: "outside-root", spelling: ".." + separator + "outside", want: allDirs},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			oldParsers := knownParser
			t.Cleanup(func() { knownParser = oldParsers })

			visited := make(map[string]bool)
			knownParser = []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
				visited[path] = true
				return nil
			}}

			if err := ParseDirRec(root, []string{tc.spelling}, &parser.DomainMap{}, false); err != nil {
				t.Fatal(err)
			}
			assertVisitedDirs(t, visited, tc.want)
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
	events := make([]string, 0, 5)
	knownParser = []ParseDirFunc{
		func(path, _ string, _ *parser.DomainMap) error {
			events = append(events, "first:"+path)
			return nil
		},
		func(path, _ string, _ *parser.DomainMap) error {
			events = append(events, "second:"+path)
			if path == child {
				return stop
			}
			return nil
		},
		func(path, _ string, _ *parser.DomainMap) error {
			events = append(events, "third:"+path)
			return nil
		},
	}

	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err != stop {
		t.Fatalf("ParseDirRec error = %v, want %v", err, stop)
	}
	want := []string{
		"first:" + root,
		"second:" + root,
		"third:" + root,
		"first:" + child,
		"second:" + child,
	}
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
	calls := make([]string, 0, 1)
	knownParsers := []ParseDirFunc{func(path, _ string, _ *parser.DomainMap) error {
		calls = append(calls, path)
		if path == root {
			return stop
		}
		return nil
	}}
	knownParser = knownParsers

	if err := ParseDirRec(root, nil, &parser.DomainMap{}, false); err != stop {
		t.Fatalf("ParseDirRec error = %v, want %v", err, stop)
	}
	if len(calls) != 1 || calls[0] != root {
		t.Fatalf("callback calls after root failure = %v, want [%s]", calls, root)
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
	inside := filepath.Join(root, "inside")
	if err := os.MkdirAll(inside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outside, "child"), 0755); err != nil {
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
	assertVisitedDirs(t, visited, []string{root, inside})
}
func FuzzParseDirRecExcludeNormalization(f *testing.F) {
	root := f.TempDir()
	target := filepath.Join(root, "target")
	targetChild := filepath.Join(target, "child")
	targeted := filepath.Join(root, "targeted")
	targetedChild := filepath.Join(targeted, "child")
	for _, path := range []string{targetChild, targetedChild} {
		if err := os.MkdirAll(path, 0755); err != nil {
			f.Fatal(err)
		}
	}

	separator := string(filepath.Separator)
	f.Add("target")
	f.Add("." + separator + "target")
	f.Add("keep" + separator + ".." + separator + "target")
	f.Add(target)
	f.Add(filepath.Join(root, "keep") + separator + ".." + separator + "target")
	f.Add(".." + separator + filepath.Base(root) + separator + "target")
	f.Add("")
	f.Add("\x00")

	oldParsers := knownParser
	f.Cleanup(func() { knownParser = oldParsers })

	allowed := map[string]struct{}{
		root:          {},
		target:        {},
		targetChild:   {},
		targeted:      {},
		targetedChild: {},
	}
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
		if !visited[root] {
			t.Fatalf("ParseDirRec(%q) did not visit traversal root", spelling)
		}
		for path := range visited {
			if _, ok := allowed[path]; !ok {
				t.Fatalf("ParseDirRec(%q) visited unexpected path %s", spelling, path)
			}
		}
	})
}
