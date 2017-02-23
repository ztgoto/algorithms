package path

import (
	"regexp"
	"strings"
)

var (
	default_path_separator  string = "/"
	cache_turnoff_threshold int    = 65536
	variable_pattern, _            = regexp.Compile(`\{[^/]+?\}`)
	wildcard_chars          []rune = []rune{'*', '?', '{'}
	trimTokens              bool   = false
	caseSensitive           bool   = true
)

// func Match(pattern, path string, fullMatch bool, uriTemplateVariables [string]string) bool {
// 	if strings.HasPrefix(path, default_path_separator) != strings.HasPrefix(pattern, default_path_separator) {
// 		return false
// 	}

// 	pattDirs := tokenizePattern(pattern)

// 	if fullMatch && caseSensitive && !isPotentialMatch(path, pattDirs) {
// 		return false
// 	}

// }

func tokenizePattern(pattern string) []string {
	return strings.Split(pattern, default_path_separator)
}

func isPotentialMatch(path string, pattDirs []string) bool {
	if !trimTokens {

		pos := 0

		for _, pattDir := range pattDirs {
			skipped := skipSeparator(path, pos, default_path_separator)
			pos += skipped
			skipped = skipSegment(path, pos, pattDir)
			if skipped < len(pattDir) {
				if skipped > 0 {
					return true
				}
				return len(pattDir) > 0 && isWildcardChar(rune(pattDir[0]))
			}
			pos += skipped
		}

	}
	return true
}

func skipSeparator(path string, pos int, separator string) int {
	skipped := 0

	sepl := len(separator)

	plen := len(path)

	begin, end := pos, pos+sepl

	for begin < plen && end < plen && path[begin:end] == separator {
		skipped += sepl
		begin += sepl
		end += sepl
	}

	return skipped
}

func skipSegment(chars string, pos int, prefix string) int {
	skipped := 0

	for _, c := range prefix {
		if isWildcardChar(c) {
			return skipped
		} else if pos+skipped >= len(chars) {
			return 0
		} else if rune(chars[pos+skipped]) == c {
			skipped++
		}
	}

	return skipped
}

func isWildcardChar(c rune) bool {
	for _, candidate := range wildcard_chars {
		if c == candidate {
			return true
		}
	}
	return false
}
