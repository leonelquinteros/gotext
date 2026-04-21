package dir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

func TestAddParser(t *testing.T) {
	initialCount := len(knownParser)
	AddParser(func(filePath, basePath string, data *parser.DomainMap) error {
		return nil
	})
	if len(knownParser) != initialCount+1 {
		t.Error("AddParser failed to add a parser")
	}
}

func TestParseDir(t *testing.T) {
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
	defer os.RemoveAll(tmpDir)

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

	dm := &parser.DomainMap{}
	callCount := 0
	AddParser(func(dirPath, basePath string, data *parser.DomainMap) error {
		callCount++
		return nil
	})

	// Reset knownParser to only have our test parser for predictable count
	oldParsers := knownParser
	knownParser = []ParseDirFunc{func(dirPath, basePath string, data *parser.DomainMap) error {
		callCount++
		return nil
	}}
	defer func() { knownParser = oldParsers }()

	err = ParseDirRec(tmpDir, []string{"exclude"}, dm, true)
	if err != nil {
		t.Errorf("ParseDirRec failed: %v", err)
	}

	// Should be called for tmpDir and subDir, but not excludeDir
	// Total 2 calls
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestGoFile_InspectFile_Coverage(t *testing.T) {
	// Minimal coverage for the InspectFile switch
	g := &GoFile{}
	g.InspectFile(nil) // Default case
}
