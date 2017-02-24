package path

import (
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
