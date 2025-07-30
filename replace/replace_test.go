package replace

import (
	"strconv"
	"strings"
	"testing"

	"golang.org/x/text/transform"
	"gotest.tools/v3/assert"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		in, out string
		tr      transform.Transformer
	}{
		{
			in: "test", out: "test",
			tr: String("", "x", 0),
		},
		{
			in: "a", out: "b",
			tr: String("a", "b", 0),
		},
		{
			in: "yes", out: "no",
			tr: String("yes", "no", 0),
		},
		{
			in: "what what what", out: "wut wut wut",
			tr: String("what", "wut", 0),
		},
		{
			in: "what what what", out: "wut what what",
			tr: String("what", "wut", 1),
		},
		{
			in: "???????", out: "*******",
			tr: String("?", "*", 0),
		},
		{
			in: "no matches", out: "no matches",
			tr: String("x", "y", 0),
		},
		{
			in: "hello", out: "heLLo",
			tr: String("l", "L", 0),
		},
		{
			in: "hello", out: "hello",
			tr: String("x", "X", 0),
		},
		{
			in: "", out: "",
			tr: String("x", "X", 0),
		},
		{
			in: "radar", out: "<r>ada<r>",
			tr: String("r", "<r>", 0),
		},
		{
			in: "banana", out: "b<>n<>n<>",
			tr: String("a", "<>", 0),
		},
		{
			in: "banana", out: "b<><>a",
			tr: String("an", "<>", 0),
		},
		{
			in: "banana", out: "b<>na",
			tr: String("ana", "<>", 0),
		},
		{
			in: "banana", out: "banana",
			tr: String("a", "a", 0),
		},
		{
			in: "xxx", out: "",
			tr: String("x", "", 0),
		},
		{
			in: strings.Repeat("foo_", 8<<10), out: strings.Repeat("bar_", 8<<10),
			tr: String("foo", "bar", 0),
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _, err := transform.String(tt.tr, tt.in)
			assert.NilError(t, err)
			assert.DeepEqual(t, result, tt.out)
		})
	}
}

func TestOverflowDst(t *testing.T) {
	s := strings.Repeat("x", 8<<10)
	tr := String("x", s, 0)

	result, _, err := transform.String(tr, "x")
	assert.NilError(t, err)
	assert.Equal(t, result, s)
}

func TestOverflowDstWithMaxCount(t *testing.T) {
	s := strings.Repeat("x", 8<<10)
	tr := String("x", s, 1)

	result, _, err := transform.String(tr, "x x")
	assert.NilError(t, err)
	assert.Equal(t, result, s+" x")
}
