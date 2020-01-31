package blog

import (
	"fmt"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/stretchr/testify/assert"
)

func TestMD(t *testing.T) {
	output := fmt.Sprintf("%s", markdown.ToHTML([]byte(`
this is a list:

* foo
* bar

ok?
`), nil, nil))
	assert.Contains(t, output, "<li>")
}

func TestMarkdowner(t *testing.T) {
	output := fmt.Sprintf("%s", markDowner(`
this is a list:

* foo
* bar

ok?
`))
	assert.Contains(t, output, "<li>")
}

func TestPostMardowning(t *testing.T) {
	post := NewPost(PostOpts{
		Body: `
this is a list:

* foo
* bar

ok?
`,
	})
	output := fmt.Sprintf("%s", markDowner(post.Body))
	assert.Contains(t, output, "<li>")
	output = fmt.Sprintf("%s", hashtagger(output))
	assert.Contains(t, output, "<li>")
}
