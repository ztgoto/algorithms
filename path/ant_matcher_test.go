package path

import (
	"testing"
)

func TestSkipSeparator(t *testing.T) {
	url := "/aa/bb/cc"
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
