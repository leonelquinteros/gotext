package dir

import (
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

func TestGoFile_InspectFile_Coverage(t *testing.T) {
	// Minimal coverage for the InspectFile switch
	g := &GoFile{}
	g.InspectFile(nil) // Default case
}
