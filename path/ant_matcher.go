package path

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var (
	path_separator          string = "/"
	cache_turnoff_threshold int    = 65536
	variable_pattern, _            = regexp.Compile(`\{[^/]+?\}`)
	wildcard_chars          []rune = []rune{'*', '?', '{'}
	trimTokens              bool   = false
	caseSensitive           bool   = true
)

func Match(pattern, path string) bool {
	return DoMatch(pattern, path, true, nil)
}

func DoMatch(pattern, path string, fullMatch bool, uriTemplateVariables map[string]string) bool {
	if strings.HasPrefix(path, path_separator) != strings.HasPrefix(pattern, path_separator) {
		return false
	}

	pattDirs := tokenizePath(pattern, path_separator, trimTokens, true)

	if fullMatch && caseSensitive && !isPotentialMatch(path, pattDirs) {
		return false
	}

	pathDirs := tokenizePath(path, path_separator, trimTokens, true)

	pattIdxStart := 0
	pattIdxEnd := len(pattDirs) - 1
	pathIdxStart := 0
	pathIdxEnd := len(pathDirs) - 1

	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxStart]
		if "**" == pattDir {
			break
		}
		if !matchStrings(pattDir, pathDirs[pathIdxStart], uriTemplateVariables) {
			return false
		}
		pattIdxStart++
		pathIdxStart++
	}

	if pathIdxStart > pathIdxEnd {
		// Path is exhausted, only match if rest of pattern is * or **'s
		if pattIdxStart > pattIdxEnd {
			if strings.HasSuffix(pattern, path_separator) {
				return strings.HasSuffix(path, path_separator)
			} else {
				return !strings.HasSuffix(path, path_separator)
			}

		}
		if !fullMatch {
			return true
		}
		if pattIdxStart == pattIdxEnd && pattDirs[pattIdxStart] == "*" && strings.HasSuffix(path, path_separator) {
			return true
		}
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if !(pattDirs[i] == "**") {
				return false
			}
		}
		return true
	} else if pattIdxStart > pattIdxEnd {
		// String not exhausted, but pattern is. Failure.
		return false
	} else if !fullMatch && "**" == pattDirs[pattIdxStart] {
		// Path start definitely matches due to "**" part in pattern.
		return true
	}

	for pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
		pattDir := pattDirs[pattIdxEnd]
		if pattDir == "**" {
			break
		}
		if !matchStrings(pattDir, pathDirs[pathIdxEnd], uriTemplateVariables) {
			return false
		}
		pattIdxEnd--
		pathIdxEnd--
	}

	if pathIdxStart > pathIdxEnd {
		// String is exhausted
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if !(pattDirs[i] == "**") {
				return false
			}
		}
		return true
	}

	for pattIdxStart != pattIdxEnd && pathIdxStart <= pathIdxEnd {
		patIdxTmp := -1
		for i := pattIdxStart + 1; i <= pattIdxEnd; i++ {
			if pattDirs[i] == "**" {
				patIdxTmp = i
				break
			}
		}
		if patIdxTmp == pattIdxStart+1 {
			// '**/**' situation, so skip one
			pattIdxStart++
			continue
		}
		// Find the pattern between padIdxStart & padIdxTmp in str between
		// strIdxStart & strIdxEnd
		patLength := patIdxTmp - pattIdxStart - 1
		strLength := pathIdxEnd - pathIdxStart + 1
		foundIdx := -1

	strLoop:
		for i := 0; i <= strLength-patLength; i++ {
			for j := 0; j < patLength; j++ {
				subPat := pattDirs[pattIdxStart+j+1]
				subStr := pathDirs[pathIdxStart+i+j]
				if !matchStrings(subPat, subStr, uriTemplateVariables) {
					continue strLoop
				}
			}
			foundIdx = pathIdxStart + i
			break
		}

		if foundIdx == -1 {
			return false
		}

		pattIdxStart = patIdxTmp
		pathIdxStart = foundIdx + patLength
	}

	for i := pattIdxStart; i <= pattIdxEnd; i++ {
		if !(pattDirs[i] == "**") {
			return false
		}
	}

	return true

}

func matchStrings(pattern, str string, uriTemplateVariables map[string]string) bool {
	stringMathcer := newStringMatcher(pattern, false)
	result, _ := stringMathcer.MatchStrings(str, uriTemplateVariables)
	return result
}

func tokenizePath(str, delimiters string, trimTokens, ignoreEmptyTokens bool) []string {

	st := strings.Split(str, path_separator)

	if !trimTokens && !ignoreEmptyTokens {
		return st
	}

	data := make([]string, 0, len(st))
	for _, v := range st {

		if trimTokens {
			v = strings.TrimSpace(v)
		}

		if !ignoreEmptyTokens || len(v) > 0 {
			data = append(data, v)
		}
	}

	return data
}

func isPotentialMatch(path string, pattDirs []string) bool {
	if !trimTokens {

		pos := 0

		for _, pattDir := range pattDirs {
			skipped := skipSeparator(path, pos, path_separator)
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
