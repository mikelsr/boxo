package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathParsing(t *testing.T) {
	cases := map[string]bool{
		"/ipfs/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n":             true,
		"/ipfs/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a":           true,
		"/ipfs/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a/b/c/d/e/f": true,
		"/ipld/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n":             true,
		"/ipld/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a":           true,
		"/ipld/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a/b/c/d/e/f": true,
		"/ipns/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a/b/c/d/e/f": true,
		"/ipns/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n":             true,
		"QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a/b/c/d/e/f":       true,
		"QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n":                   true,
		"/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n":                  false,
		"/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n/a":                false,
		"/ipfs/foo": false,
		"/ipfs/":    false,
		"ipfs/":     false,
		"ipfs/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n": false,
		"/ipld/foo": false,
		"/ipld/":    false,
		"ipld/":     false,
		"ipld/QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n": false,
		"/ipns":            false,
		"/ipns/domain.net": true,
	}

	for p, expected := range cases {
		_, err := NewPath(p)
		valid := err == nil
		assert.Equal(t, expected, valid, "expected %s to have valid == %t", p, expected)
	}
}

func TestNoComponents(t *testing.T) {
	for _, s := range []string{
		"/ipfs/",
		"/ipns/",
		"/ipld/",
	} {
		_, err := NewPath(s)
		assert.ErrorContains(t, err, "not enough path components")
	}
}

func TestInvalidPaths(t *testing.T) {
	for _, s := range []string{
		"/ipfs",
		"/testfs",
		"/",
	} {
		_, err := NewPath(s)
		assert.ErrorContains(t, err, "not enough path components")
	}
}
