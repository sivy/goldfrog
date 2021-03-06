package blog

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// const (
// 	testDb string = "../../tests/data/test.db"
// )
var configStr = `
blog:
  title: monkinetic.blog
  subhead: Since 1999, XVI Edition
  url: "http://monkinetic.blog"
  author:
    name: Steve Ivy
    email: steveivy@gmail.com
    image: "http://monkinetic.blog/static/images/sivy_avatar_256.png"
	timezone: "America/Phoenix"
`
var CONFIG = LoadConfigStr(configStr)

/*
Do-nothing file saver
*/
type NullPostsRepo struct{}

func (npr *NullPostsRepo) ListPostFiles() []string {
	return []string{}
}
func (npr *NullPostsRepo) SavePostFile(post *Post) error {
	return nil
}
func (npr *NullPostsRepo) DeletePostFile(post *Post) error {
	return nil
}

func TestCreatePostHandlerNote(t *testing.T) {
	// Setup
	data := "title=foo&slug=foo&body=note #content"

	req, err := http.NewRequest("POST", "/new", strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	err = initDb(testDb)
	assert.Nil(t, err)
	defer func() {
		// Teardown
		os.Remove(testDb)
	}()

	db, err := GetDb(testDb)
	assert.NotNil(t, db)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := CreateNewPostFunc(CONFIG, db, &NullPostsRepo{})

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusFound, rr.Code)
}
