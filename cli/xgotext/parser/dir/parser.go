package dir

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

// ParseDirFunc parses one directory
type ParseDirFunc func(filePath, basePath string, data *parser.DomainMap) error

var knownParser []ParseDirFunc

// AddParser to the known parser list
func AddParser(parser ParseDirFunc) {
	if knownParser == nil {
		knownParser = []ParseDirFunc{parser}
	} else {
		knownParser = append(knownParser, parser)
	}
}

// ParseDir calls all known parser for each directory
func ParseDir(dirPath, basePath string, data *parser.DomainMap) error {
	var err error
	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	basePath, err = filepath.Abs(basePath)
	if err != nil {
		return err
	}

	for _, parser := range knownParser {
		err := parser(dirPath, basePath, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseDirRec calls all known parser for each directory
func ParseDirRec(dirPath string, exclude []string, data *parser.DomainMap, verbose bool) error {
	var err error
	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath = filepath.Clean(dirPath)

	excludeDirs := make([]string, 0, len(exclude))
	for _, excludeDir := range exclude {
		if !filepath.IsAbs(excludeDir) {
			excludeDir = filepath.Join(dirPath, excludeDir)
		}
		excludeDir = filepath.Clean(excludeDir)

		// The traversal root is always visited. Exclusions outside the root
		// cannot match any path WalkDir visits and are ignored.
		if excludeDir == dirPath {
			continue
		}
		rel, relErr := filepath.Rel(dirPath, excludeDir)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		excludeDirs = append(excludeDirs, excludeDir)
	}

	return filepath.WalkDir(dirPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
			for _, excludeDir := range excludeDirs {
				if path == excludeDir || strings.HasPrefix(path, excludeDir+string(filepath.Separator)) {
					return filepath.SkipDir
				}
			}

			if verbose {
				log.Print(path)
			}

			if err := ParseDir(path, dirPath, data); err != nil {
				return err
			}
		}
		return nil
	})
}
