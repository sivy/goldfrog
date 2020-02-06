package syndication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripHTML(t *testing.T) {
	content := "<p>foo</p>"

	content = stripHTML(content)

	assert.NotContains(t, content, "<p>")
}
