package path

type PathMatcher interface {
	IsPattern(string) bool

	Match(string, string) bool

	MatchStart(string, string) bool

	ExtractPathWithinPattern(string, string) string

	ExtractUriTemplateVariables(string, string) map[string]string

	// GetPatternComparator(string) Comparator

	Combine(string, string) string
}
