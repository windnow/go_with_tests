package main

import (
	"reflect"
	"testing"
)

type Person struct {
	Name    string
	Profile Profile
}

type Profile struct {
	Age  int
	City string
}
type testCase struct {
	Name          string
	Input         interface{}
	ExpectedCalls []string
}

func fn(got *[]string) func(string) {
	return func(input string) {
		*got = append(*got, input)
	}
}

func TestWalk(t *testing.T) {
	cases := getCases()
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			var got []string
			walk(test.Input, fn(&got))

			if !reflect.DeepEqual(got, test.ExpectedCalls) {
				t.Errorf("got %v, want %v", got, test.ExpectedCalls)
			}
		})
	}

	t.Run("with maps", func(t *testing.T) {
		aMap := map[string]string{
			"Foo": "Bar",
			"Baz": "Boz",
		}

		var got []string
		walk(aMap, fn(&got))

		assertContains(t, got, "Bar")
		assertContains(t, got, "Boz")
	})

	t.Run("with channels", func(t *testing.T) {
		aChannel := make(chan Profile)

		go func() {
			aChannel <- Profile{33, "Berlin"}
			aChannel <- Profile{34, "Aktobe"}
			close(aChannel)
		}()
		var got []string
		want := []string{"Berlin", "Aktobe"}

		walk(aChannel, fn(&got))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with function", func(t *testing.T) {
		aFunction := func() (Profile, Profile) {
			return Profile{33, "Berlin"}, Profile{34, "Aktobe"}
		}

		var got []string
		want := []string{"Berlin", "Aktobe"}

		walk(aFunction, fn(&got))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func assertContains(t *testing.T, haystack []string, needle string) {
	t.Helper()
	contains := false
	for _, x := range haystack {
		if x == needle {
			contains = true
			break
		}
	}
	if !contains {
		t.Errorf("expected %+v to contain %q but it didn't", haystack, needle)
	}
}

func getCases() []testCase {
	return []testCase{
		{
			"Struct with one string field",
			struct {
				Name string
			}{"Chris"},
			[]string{"Chris"},
		},
		{
			"Struct with two string fields",
			struct {
				Name string
				Sity string
			}{
				"Chris",
				"London",
			},
			[]string{"Chris", "London"},
		},
		{
			"Struct with non string fields",
			struct {
				Name string
				Age  int
			}{"Chris", 33},
			[]string{"Chris"},
		}, {
			"Nested fields",
			Person{
				"Chris",
				Profile{
					33,
					"London",
				},
			},
			[]string{"Chris", "London"},
		},
		{
			"Pointers to thing",
			&Person{
				"Chris",
				Profile{
					33,
					"London",
				},
			},
			[]string{"Chris", "London"},
		},
		{
			"Slices",
			[]Profile{
				{33, "London"},
				{34, "Reykjavik"},
			},
			[]string{"London", "Reykjavik"},
		},
		{
			"Arrays",
			[2]Profile{
				{33, "London"},
				{37, "Astana"},
			},
			[]string{"London", "Astana"},
		},
	}

}
