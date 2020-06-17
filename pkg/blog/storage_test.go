package blog

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	author_tz, _ := time.LoadLocation("America/Phoenix")
	initDb(testDb)
	db, _ := GetDb(testDb)
	defer func() {
		os.Remove(testDb)
	}()
	dbs := NewSqliteStorage(db)

	author_time := time.Now().In(author_tz)
	utc_time := author_time.UTC()

	p := NewPost(PostOpts{
		Title: "the title",
		Slug:  "the-title",
		Tags:  []string{"tag"},
		FrontMatter: map[string]string{
			"twitter_url":  "twitter url",
			"mastodon_url": "mastodon url",
		},
		PostDate: utc_time,
	})
	logger.Infof("p: %v", p)

	err := dbs.CreatePost(&p)
	assert.Nil(t, err)

	// Make sure loading the post gets the same information
	p2, err := dbs.GetPostBySlug("the-title")
	logger.Infof("p2: %v", p2)
	assert.Nil(t, err)

	assert.NotNil(t, p2)

	assert.Equal(t, p.Title, p2.Title)
	assert.Equal(t, p.Slug, p2.Slug)
	assert.Equal(t, p.Tags, p2.Tags)
	assert.Equal(
		t,
		p.PostDate.Format(time.RFC3339),
		p2.PostDate.Format(time.RFC3339))

	logger.Infof(
		"p frontmatter: %v || p2 frontmatter: %v",
		p.FrontMatter, p2.FrontMatter)
	assert.Equal(t, p.FrontMatter, p2.FrontMatter)

	// make a change and save
	p2.FrontMatter["featured_image"] = "My Uploaded Image.jpg"
	// Make sure the change persisted
	err = dbs.SavePost(p2)
	assert.Nil(t, err)

	p3, err := dbs.GetPostBySlug("the-title")
	logger.Infof("p3: %v", p3)
	assert.Equal(t, p2.Title, p3.Title)
	assert.Equal(t, p2.Slug, p3.Slug)
	assert.Equal(t, p2.Tags, p3.Tags)
	assert.Equal(t, p2.FrontMatter, p3.FrontMatter)
	assert.Equal(
		t,
		p2.PostDate.Format(time.RFC3339),
		p3.PostDate.Format(time.RFC3339))
}
