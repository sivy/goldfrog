package blog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPostToString(t *testing.T) {
	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
		Body: "post body",
	})
	assert.NotNil(t, p)
	postStr := p.ToString()

	assert.Contains(t, postStr, "title: the title")
	assert.Contains(t, postStr, "slug: the-title")
	assert.Contains(t, postStr, "tags: tag")
	assert.Contains(t, postStr, "twitter_url: twitter url")
	assert.Contains(t, postStr, "mastodon_url: mastodon url")
	assert.Contains(t, postStr, "---")
	assert.Contains(t, postStr, "post body")

}

func TestFrontMatterYaml(t *testing.T) {
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
	fmYamlStr := p.FrontMatterYAML()

	assert.Contains(t, fmYamlStr, "title: the title")
	assert.Contains(t, fmYamlStr, "slug: the-title")
	assert.Contains(t, fmYamlStr, "tags: tag")
	assert.Contains(t, fmYamlStr, "twitter_url: twitter url")
	assert.Contains(t, fmYamlStr, "mastodon_url: mastodon url")
}

func TestPostPermaShortId(t *testing.T) {
	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
		Body: "post body",
	})
	assert.NotNil(t, p)
	assert.Equal(t, p.PermaShortId(), "the-title")

}

func TestPostPermaLink(t *testing.T) {
	postdate, _ := time.Parse("2006-01-02", "2006-01-02")
	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
		Body:     "post body",
		PostDate: postdate,
	})
	assert.NotNil(t, p)
	assert.Equal(t, p.PermaLink(), "/2006/01/02/the-title")
}

func TestPostTagString(t *testing.T) {
	p := NewPost(PostOpts{
		Tags: []string{"tag", "string"},
	})
	assert.NotNil(t, p)
	assert.Equal(t, p.TagString(), "tag, string")
}
