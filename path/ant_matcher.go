package path

import (
	"bytes"
	"errors"
	"fmt"
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

// 	pathDirs := tokenizePath(path)

// 	pattIdxStart := 0
// 	pattIdxEnd := len(pattDirs) - 1
// 	pathIdxStart := 0
// 	pathIdxEnd := len(pattDirs) - 1

// 	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
// 		pattDir := pattDirs[pattIdxStart]
// 		if "**" == pattDir {
// 			break
// 		}
// 		if !matchStrings(pattDir, pathDirs[pathIdxStart], uriTemplateVariables) {
// 			return false
// 		}
// 		pattIdxStart++
// 		pathIdxStart++
// 	}

// 	if pathIdxStart > pathIdxEnd {
// 			// Path is exhausted, only match if rest of pattern is * or **'s
// 			if pattIdxStart > pattIdxEnd {
// 				return (pattern.endsWith(this.pathSeparator) ? path.endsWith(this.pathSeparator) :
// 						!path.endsWith(this.pathSeparator));
// 			}
// 			if (!fullMatch) {
// 				return true;
// 			}
// 			if (pattIdxStart == pattIdxEnd && pattDirs[pattIdxStart].equals("*") && path.endsWith(this.pathSeparator)) {
// 				return true;
// 			}
// 			for (int i = pattIdxStart; i <= pattIdxEnd; i++) {
// 				if (!pattDirs[i].equals("**")) {
// 					return false;
// 				}
// 			}
// 			return true;
// 		}

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

func quote(s string, start, end int) string {
	if start == end {
		return ""
	}
	return RegexpQuote(s[start:end])
}

func RegexpQuote(s string) string {
	slashEIndex := strings.Index(s, "\\E")
	if slashEIndex == -1 {
		return "\\Q" + s + "\\E"
	}

	var sb bytes.Buffer
	sb.WriteString("\\Q")

	slashEIndex = 0
	current := 0

	for true {

		slashEIndex = IndexFrom(s, "\\E", current)

		if slashEIndex == -1 {
			break
		}

		sb.WriteString(s[current:slashEIndex])
		current = slashEIndex + 2
		sb.WriteString("\\E\\\\E\\Q")
	}
	sb.WriteString(s[current:])
	sb.WriteString("\\E")
	return sb.String()
}

func IndexFrom(s, sep string, from int) int {
	n := len(sep)
	if n == 0 {
		return 0
	}
	c := sep[0]
	if n == 1 {
		// special case worth making fast
		for i := from; i < len(s); i++ {
			if s[i] == c {
				return i
			}
		}
		return -1
	}
	// n > 1
	for i := from; i+n <= len(s); i++ {
		if s[i] == c && s[i:i+n] == sep {
			return i
		}
	}
	return -1
}

type AntPathStringMatcher struct {
	glob_pattern             *regexp.Regexp
	pattern                  *regexp.Regexp
	default_variable_pattern string
	variableNames            []string
}

func (this *AntPathStringMatcher) MatchStrings(str string, uriTemplateVariables map[string]string) (bool, error) {

	fmt.Println(this.pattern.String())
	fmt.Println(this.variableNames)

	if this.pattern.MatchString(str) {
		if uriTemplateVariables != nil {
			matches := this.pattern.FindStringSubmatch(str)
			if len(this.variableNames) != len(matches)-1 {
				return false, errors.New("The number of capturing groups in the pattern segment " +
					this.pattern.String() + " does not match the number of URI template variables it defines, " +
					"which can occur if capturing groups are used in a URI template regex. " +
					"Use non-capturing groups instead.")
			}
			for i := 1; i < len(matches); i++ {
				name := this.variableNames[i-1]
				value := matches[i]
				uriTemplateVariables[name] = value
			}
		}
		return true, nil
	} else {
		return false, nil
	}

}

func newStringMatcher(pattern string, caseSensitive bool) *AntPathStringMatcher {
	stringMatcher := new(AntPathStringMatcher)
	stringMatcher.glob_pattern, _ = regexp.Compile(`\?|\*|\{((?:\{[^/]+?\}|[^/{}]|\\[{}])+?)\}`)
	stringMatcher.default_variable_pattern = "(.*)"

	var b bytes.Buffer

	matches := stringMatcher.glob_pattern.FindAllStringSubmatch(pattern, -1)
	indexs := stringMatcher.glob_pattern.FindAllStringSubmatchIndex(pattern, -1)
	end := 0
	for i, submatch := range matches {

		b.WriteString(quote(pattern, end, indexs[i][0]))

		match := submatch[0]

		if "?" == match {
			b.WriteString(".")
		} else if "*" == match {
			b.WriteString(".*")
		} else if strings.HasPrefix(match, "{") && strings.HasSuffix(match, "}") {
			colonIdx := strings.Index(match, ":")
			if colonIdx == -1 {
				b.WriteString(stringMatcher.default_variable_pattern)
				stringMatcher.variableNames = append(stringMatcher.variableNames, submatch[1])
			} else {

				variablePattern := match[colonIdx+1 : len(match)-1]

				b.WriteString("(")
				b.WriteString(variablePattern)
				b.WriteString(")")
				stringMatcher.variableNames = append(stringMatcher.variableNames, match[1:colonIdx])
			}

		}
		end = indexs[i][1]

	}

	b.WriteString(quote(pattern, end, len(pattern)))

	if caseSensitive {
		stringMatcher.pattern, _ = regexp.Compile("^" + b.String() + "$")
	} else {
		stringMatcher.pattern, _ = regexp.Compile("(?i:^" + b.String() + "$)")
	}

	return stringMatcher
}
