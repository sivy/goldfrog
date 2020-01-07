package blog

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncHashTags(t *testing.T) {
	s := `this is a [post](#foo) about #testing and #golang`
	res := hashtagger(s)
	fmt.Printf("res: %v", res)
	assert.NotEmpty(t, res)

	s = "(#DIY-ing)"
	res = hashtagger(s)
	fmt.Printf("res: %v", res)
	assert.NotEmpty(t, res)

}

func TestMarkdownHashTags(t *testing.T) {
	s := `this is a [post](#foo) about #testing and #golang`
	html := markDowner(s)

	res := hashtagger(html)

	fmt.Printf("res: %v", res)
	assert.NotEmpty(t, res)

	s = "(#DIY-ing)"
	res = hashtagger(s)
	fmt.Printf("res: %v", res)
	assert.NotEmpty(t, res)

}

func TestTagSearch(t *testing.T) {
	tags := "foo, bar,baz"
	tag := "bar"

	re := regexp.MustCompile(fmt.Sprintf("(?:\\A|[\\W])(%s)(?:\\W|\\z)", tag))
	assert.True(t, re.MatchString("bar"))

	tag = "foo"
	re = regexp.MustCompile(fmt.Sprintf("(?:\\A|[\\W])(%s)(?:\\W|\\z)", tag))
	assert.True(t, re.MatchString(tags))

	tag = "bar"
	re = regexp.MustCompile(fmt.Sprintf("(?:\\A|[\\W])(%s)(?:\\W|\\z)", tag))
	assert.True(t, re.MatchString(tags))

	tag = "baz"

}
