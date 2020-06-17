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

func TestCreatePostHandlerNote(t *testing.T) {
	// Setup
	data := "title=foo&slug=foo&body=note #content"

	req, err := http.NewRequest("POST", "/new", strings.NewReader(data))
	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

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
	dbs := NewMockDBStorage()

	var CONFIG = LoadConfig(testConfigDir)

	rr := httptest.NewRecorder()
	handler := CreateNewPostFunc(CONFIG, dbs, &MockPostsRepo{})

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusFound, rr.Code)
}

// TODO: TestCreatePostHandlerPost
