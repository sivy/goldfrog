package blog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Log(err)
		assert.Fail(t, "Could not get cwd")
	}
	path := filepath.Join(
		cwd,
		"../../tests/data/2019-12-30-test-post.md")

	post, err := ParseFile(path)

	assert.NotNil(t, post)
	assert.Nil(t, err)

	assert.Equal(t, "test post", post.Title)
	for _, tag := range post.Tags {
		assert.Contains(t, []string{"test", "post", "hashtag"}, tag)
	}
	assert.Equal(t, "Post body #hashtag", post.Body)

	assert.Equal(t, "2019-12-30T22:24:00Z", post.PostDate.Format(POSTTIMESTAMPFMT))

	assert.IsType(t, make(map[string]string), post.FrontMatter)
	assert.Contains(t, post.FrontMatter, "twitter_id")
	assert.Contains(t, post.FrontMatter, "mastodon_id")
	assert.Contains(t, post.FrontMatter, "goodreads_id")

	assert.Equal(t, "123", post.FrontMatter["twitter_id"])
	assert.Equal(t, "abc", post.FrontMatter["mastodon_id"])
	assert.Equal(t, "def", post.FrontMatter["goodreads_id"])

	assert.Equal(t, "/2019/12/30/test-post", post.PermaLink())

}

func TestGetDateWithGoodDateStr(t *testing.T) {

	dt, err := getPostDate(
		"2019-12-31T11:59:59Z", "2020-01-01-happy-new-years.md")

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
	slug := MakePostSlug(title)
	assert.Equal(t, "this-is-a-test", slug)

	title = "this isn't a test"
	slug = MakePostSlug(title)
	assert.Equal(t, "this-isnt-a-test", slug)

	title = "this is 1 test"
	slug = MakePostSlug(title)
	assert.Equal(t, "this-is-1-test", slug)

	title = "this is 1 test?"
	slug = MakePostSlug(title)
	assert.Equal(t, "this-is-1-test", slug)
}

func TestGetHashTags(t *testing.T) {
	s := `this is a post about #testing and #golang`
	res := GetHashTags(s)
	assert.NotEmpty(t, res)
	assert.Equal(t, []string{"testing", "golang"}, res)

	res = GetHashTags("")
	assert.Empty(t, res)
}
