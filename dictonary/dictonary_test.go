package dictonary

import "testing"

func TestSearch(t *testing.T) {
	dictonary := Dictonary{"test": "this is just a test"}

	t.Run("known word", func(t *testing.T) {
		got, _ := dictonary.Search("test")
		want := "this is just a test"

		assertStrings(t, got, want)
	})

	t.Run("unknown word", func(t *testing.T) {
		_, got := dictonary.Search("unknown")

		assertError(t, got, ErrNotFound)
	})
}

func TestAdd(t *testing.T) {
	dictonary := Dictonary{}
	dictonary.Add("test", "this is just a test")

	want := "this is just a test"
	got, err := dictonary.Search("test")
	if err != nil {
		t.Fatal("should be added word:", err)
	}

	assertStrings(t, got, want)
}

func assertStrings(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func assertError(t *testing.T, got, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got error %q want %q", got, want)
	}
}
