package source

import (
	"path/filepath"
	"sort"
	"strings"
)

type sourceGlobInterface interface {
	IsDir(dir string) (isDir bool, err error)
	Readdirnames(dir string) ([]string, error)
}

func sourceGlob(g sourceGlobInterface, pattern string) (matches []string, err error) {
	return sourceGlobWithLimit(g, pattern, 0)
}

func sourceGlobWithLimit(g sourceGlobInterface, pattern string, depth int) (matches []string, err error) {
	// This limit is used prevent stack exhaustion issues. See CVE-2022-30632.
	const pathSeparatorsLimit = 10000
	if depth == pathSeparatorsLimit {
		return nil, filepath.ErrBadPattern
	}

	// Check pattern is well-formed.
	if _, err = filepath.Match(pattern, ""); err != nil {
		return nil, err
	}
	if !sourceGlobHas(pattern) {
		return []string{pattern}, nil
	}

	dir, file := filepath.Split(pattern)
	dir = cleanGlobPath(dir)

	if !sourceGlobHas(dir) {
		return sourceGlobDir(g, dir, file, nil)
	}

	// Prevent infinite recursion.
	if dir == pattern {
		return nil, filepath.ErrBadPattern
	}

	var m []string
	m, err = sourceGlobWithLimit(g, dir, depth+1)
	if err != nil {
		return nil, err
	}
	for _, d := range m {
		matches, err = sourceGlobDir(g, d, file, matches)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

// cleanGlobPath prepares path for glob matching.
func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	case string(filepath.Separator):
		// do nothing to the path
		return path
	default:
		return path[0 : len(path)-1] // chop off trailing separator
	}
}

func sourceGlobDir(g sourceGlobInterface, dir, pattern string, matches []string) (m []string, e error) {
	m = matches
	if isDir, err := g.IsDir(dir); err != nil || !isDir {
		return m, err
	}

	names, err := g.Readdirnames(dir)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	for _, n := range names {
		matched, err := filepath.Match(pattern, n)
		if err != nil {
			return m, err
		}
		if matched {
			m = append(m, filepath.Join(dir, n))
		}
	}
	return m, nil
}

func sourceGlobHas(path string) bool {
	return strings.ContainsAny(path, `*?[\`)
}
