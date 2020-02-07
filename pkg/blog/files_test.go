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
	assert.Equal(t, []string{"test", "post", "hashtag"}, post.Tags)
	assert.Equal(t, "Post body #hashtag", post.Body)

	assert.Equal(t, "2019-12-30 22:24", post.PostDate.Format(POSTTIMESTAMPFMT))

	assert.IsType(t, make(map[string]string), post.FrontMatter)
	assert.Contains(t, post.FrontMatter, "twitter_id")
	assert.Contains(t, post.FrontMatter, "mastodon_id")
	assert.Contains(t, post.FrontMatter, "goodreads_id")

	assert.Equal(t, "123", post.FrontMatter["twitter_id"])
	assert.Equal(t, "abc", post.FrontMatter["mastodon_id"])
	assert.Equal(t, "def", post.FrontMatter["goodreads_id"])

	assert.Equal(t, "/2019/12/30/test-post", post.PermaLink())

}

func TestPostToString(t *testing.T) {
	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
	})
	assert.NotNil(t, p)
	postStr := p.ToString()

	assert.Contains(t, postStr, "title: the title")
	assert.Contains(t, postStr, "slug: the-title")
	assert.Contains(t, postStr, "tags: tag")
	assert.Contains(t, postStr, "twitter_url: twitter url")
	assert.Contains(t, postStr, "mastodon_url: mastodon url")
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

	title = "this is 1 test?"
	slug = makePostSlug(title)
	assert.Equal(t, "this-is-1-test", slug)
}

func TestGetHashTags(t *testing.T) {
	s := `this is a post about #testing and #golang`
	res := getHashTags(s)
	assert.NotEmpty(t, res)
	assert.Equal(t, []string{"testing", "golang"}, res)

	res = getHashTags("")
	assert.Empty(t, res)
}

// func TestMicroMessage(t *testing.T) {
// 	opts := MicroMessageOpts{
// 		Title:     "Some Title",
// 		PermaLink: "http://example.com/YYYY/MM/DD/some-title",
// 		MaxLength: 280,
// 		Tags:      []string{"tag1", "tag2"},
// 	}

// 	source := `
// But I must explain to you how all this mistaken idea of denouncing pleasure and
// praising pain was born and I will give you a complete account of the system.

// And  expound the actual teachings of the great explorer of the truth, the
// master-builder of human happiness. No one rejects, dislikes, or avoids pleasure
// itself, because it is pleasure, but because those who do not know how to pursue
// pleasure rationally encounter consequences that are extremely painful.

// Nor again is there anyone who loves or pursues or desires to obtain pain of itself,
// because it is pain, but because occasionally circumstances occur in which toil and
// pain can procure him some great pleasure. To take a trivial example, which of us ever
// undertakes laborious physical exercise, except to obtain some advantage from it?

// But who has any right to find fault with a man who chooses to enjoy a pleasure that
// has no annoying consequences, or one who avoids a pain that produces no resultant
// pleasure?`

// 	output := makeMicroMessage(source, opts)
// 	assert.Contains(t, output, opts.Title)
// 	assert.Contains(t, output, opts.PermaLink)
// 	assert.Contains(t, output, "#tag1 #tag2")
// 	// assert.Nil(t, output)
// }

// func TestNoteMicroMessage(t *testing.T) {
// 	opts := MicroMessageOpts{
// 		ShortID:   "txt-abc123",
// 		MaxLength: 280,
// 		Tags:      []string{"tag1", "tag2"},
// 	}

// 	// titleLen := len(title)
// 	// linkLen := len(link)

// 	source := `
// But I must explain to you how all this mistaken idea of denouncing pleasure and
// praising pain was born and I will give you a complete account of the system.
// `

// 	output := makeMicroMessage(source, opts)
// 	assert.Contains(
// 		t, output, fmt.Sprintf("(monkinetic %s)", opts.ShortID))
// 	assert.Contains(
// 		t, output, "#tag1 #tag2")
// 	assert.Contains(
// 		t, output, strings.TrimSpace(source))
// 	// assert.Nil(t, output)
// }
