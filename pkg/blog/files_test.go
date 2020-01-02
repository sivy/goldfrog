package blog

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFrontMatterItem(t *testing.T) {
	frontmatter := `
title: blog post
tags: bar, baz
`

	title := GetFrontMatterItem(frontmatter, "title")

	assert.NotNil(t, title)
	assert.Equal(t, title, "blog post")

	tags := GetFrontMatterItem(frontmatter, "tags")
	assert.NotNil(t, tags)
	assert.Equal(t, tags, "bar, baz")

}

func TestParseFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Log(err)
		assert.Fail(t, "Could not get cwd")
	}
	path := filepath.Join(
		cwd,
		"../../tests/data/post.md")

	post, err := ParseFile(path)

	assert.NotNil(t, post)
	assert.Nil(t, err)

	assert.Equal(t, "test post", post.Title)
	assert.Equal(t, []string{"test", "post"}, post.Tags)
	assert.Equal(t, "Post body", post.Body)
}

func TestGetDateWithGoodDateStr(t *testing.T) {
	dt, err := getPostDate("2019-12-31 11:59:59", "2020-01-01-happy-new-years.md")

	assert.Nil(t, err)
	assert.NotNil(t, dt)

	assert.Equal(t, dt.Year(), 2019)
	assert.Equal(t, dt.Month(), time.Month(12))
	assert.Equal(t, dt.Day(), 31)
	assert.Equal(t, dt.Hour(), 11)
	assert.Equal(t, dt.Minute(), 59)
	assert.Equal(t, dt.Second(), 59)
}

func TestGetDateWithBadDateStr(t *testing.T) {
	dt, err := getPostDate("", "2020-01-01-happy-new-years.md")

	assert.Nil(t, err)
	assert.NotNil(t, dt)

	assert.Equal(t, 2020, dt.Year())
	assert.Equal(t, time.Month(1), dt.Month())
	assert.Equal(t, 1, dt.Day())
	assert.Equal(t, 0, dt.Hour())
	assert.Equal(t, 0, dt.Minute())
	assert.Equal(t, 0, dt.Second())
}

func TestPostSlug(t *testing.T) {
	filename := "2019-12-31-post-slug-test.md"

	slug := getPostSlugFromFile(filename)

	assert.NotNil(t, slug)
	assert.Equal(t, "post-slug-test", slug)
}

func TestMakePostSlug(t *testing.T) {
	title := "this is a test"
	slug := makePostSlug(title)
	assert.Equal(t, "this-is-a-test", slug)

	title = "this isn't a test"
	slug = makePostSlug(title)
	assert.Equal(t, "this-isnt-a-test", slug)

	title = "this is 1 test"
	slug = makePostSlug(title)
	assert.Equal(t, "this-is-1-test", slug)

}
