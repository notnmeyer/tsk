package outputformat

import (
	"testing"
)

func TestIsValid(t *testing.T) {
	want, got := true, IsValid("json")
	if want != got {
		t.Errorf("got %t, wanted %t\n", got, want)
	}

	want, got = true, IsValid("md")
	if want != got {
		t.Errorf("got %t, wanted %t\n", got, want)
	}

	want, got = true, IsValid("toml")
	if want != got {
		t.Errorf("got %t, wanted %t\n", got, want)
	}

	want, got = false, IsValid("XML")
	if want != got {
		t.Errorf("got %t, wanted %t\n", got, want)
	}
}
