package dictonary

import "errors"

// Dictonary ...
// map[string]string for store "key": "value" structure
type Dictonary map[string]string

// ErrNotFound an error instance, that will be return on search fail
var ErrNotFound = errors.New("could not find the word you were looking for")

// Search ...
// for escape d[word] syntax
func (d Dictonary) Search(word string) (string, error) {
	definition, ok := d[word]
	if !ok {
		return "", ErrNotFound
	}
	return definition, nil
}

// Add add new word to dictonary
func (d Dictonary) Add(word, definition string) {
	d[word] = definition
}
