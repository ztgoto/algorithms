package path

import (
	"fmt"
	"testing"
)

func TestSkipSeparator(t *testing.T) {
	url := "/ab/bb/cc"
	skipped := skipSeparator(url, 0, "/b")
	if skipped != 0 {
		t.Errorf("skipSeparator logic error")
	}

	skipped = skipSeparator(url, 3, "/b")
	if skipped != 2 {
		t.Errorf("skipSeparator logic error")
	}

	skipped = skipSeparator(url, 10, "/b")
	if skipped != 0 {
		t.Errorf("skipSeparator logic error")
	}
}

func TestSkipSegment(t *testing.T) {
	url := "/ab/bb/cc"
	skipped := skipSegment(url, 0, "ac")
	if skipped != 0 {
		t.Errorf("skipSegment logic error")
	}

	skipped = skipSegment(url, 1, "ac")
	if skipped != 1 {
		t.Errorf("skipSegment logic error")
	}

	skipped = skipSegment(url, 1, "abc")
	if skipped != 2 {
		t.Errorf("skipSegment logic error")
	}

	skipped = skipSegment(url, 10, "ac")
	if skipped != 0 {
		t.Errorf("skipSegment logic error")
	}

	skipped = skipSegment(url, 1, "{ab}")
	if skipped != 0 {
		t.Errorf("skipSegment logic error")
	}
}

func TestRegexpQuote(t *testing.T) {
	s := "/aa/bb/c\\Ec"
	result := RegexpQuote(s)
	if result != `\Q/aa/bb/c\E\\E\Qc\E` {
		t.Errorf("RegexpQuote logic error")
	}
}

func TestMatchStrings(t *testing.T) {
	pattern := "/aa/bb/cc/{ab}/{cd}"
	str := "/aa/bb/cc/123/456"
	stringMathcer := newStringMatcher(pattern, false)
	uriTemplateVariables := map[string]string{}
	result, _ := stringMathcer.MatchStrings(str, uriTemplateVariables)
	fmt.Println(result)
	fmt.Println(uriTemplateVariables)
}

func TestMatch(t *testing.T) {
	pattern := "/aa/bb/**/*.jsp"
	path := "/aa/bb/cc/ee/dd.jsp"
	result := Match(pattern, path)
	fmt.Println("TestMatch:", result)
}

func BenchmarkSkipSeparator(b *testing.B) {
	url := "/ab/bb/cc"
	for i := 0; i < b.N; i++ {
		skipSeparator(url, 3, "/b")
	}
}

func BenchmarkSkipSegment(b *testing.B) {
	url := "/ab/bb/cc"
	for i := 0; i < b.N; i++ {
		skipSegment(url, 1, "abc")
	}
}

func BenchmarkMatch(b *testing.B) {
	b.StopTimer()
	pattern := "/aa/bb/**/*.jsp"
	path := "/aa/bb/cc/ee/dd.jsp"
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Match(pattern, path)
	}
}
