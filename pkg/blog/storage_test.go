package blog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDb string = "../../tests/data/test.db"
)

func TestGetDb(t *testing.T) {
	db, err := GetDb(testDb)
	defer func() {
		os.Remove(testDb)
	}()

	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestInitDb(t *testing.T) {
	err := initDb(testDb)
	defer func() {
		os.Remove(testDb)
	}()

	assert.Nil(t, err)
}

func TestCreateSavePostPlain(t *testing.T) {
	initDb(testDb)
	db, _ := GetDb(testDb)

	defer func() {
		os.Remove(testDb)
	}()

	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
	})
	err := CreatePost(db, &p)
	assert.Nil(t, err)

	// Make sure loading the post gets the same information
	p2, err := GetPostBySlug(db, "the-title")
	assert.Nil(t, err)

	assert.NotNil(t, p2)

	assert.Equal(t, p.Title, p2.Title)
	assert.Equal(t, p.Slug, p2.Slug)
	assert.Equal(t, p.Tags, p2.Tags)
	assert.Equal(t, p.FrontMatter, p2.FrontMatter)

	// make a change and save
	p2.FrontMatter["featured_image"] = "My Uploaded Image.jpg"
	// Make sure the change persisted
	err = SavePost(db, p2)
	assert.Nil(t, err)

	p3, err := GetPostBySlug(db, "the-title")
	assert.Equal(t, p2.Title, p3.Title)
	assert.Equal(t, p2.Slug, p3.Slug)
	assert.Equal(t, p2.Tags, p3.Tags)
	assert.Equal(t, p2.FrontMatter, p3.FrontMatter)

}
