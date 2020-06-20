package blog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (wc *MockWebmentionClient) FindLinks(htmlStr string) ([]string, error) {
	return wc.mockedLinks, nil
}

func TestWebMentionHandler(t *testing.T) {
	// Setup
	var params = []map[string]string{
		// bad source and target the same
		{
			"source":       "",
			"sourceBody":   "",
			"target":       "",
			"responseCode": fmt.Sprintf("%d", 400),
			"responseBody": "Webmention source:  cannot be the same as the target: ",
		},
		{
			"source":       "http://someblog.com",
			"sourceBody":   "",
			"target":       "http://someblog.com",
			"responseCode": fmt.Sprintf("%d", 400),
			"responseBody": "Webmention source: http://someblog.com cannot be the same as the target: http://someblog.com",
		},
		// unparseable source
		{
			"source":       "floop",
			"sourceBody":   "",
			"target":       "",
			"responseCode": fmt.Sprintf("%d", 400),
			"responseBody": "Could not parse source: floop",
		},
		// unparseable target
		{
			"source":       "http://someblog.com",
			"sourceBody":   "",
			"target":       "floop",
			"responseCode": fmt.Sprintf("%d", 400),
			"responseBody": "Could not parse target: floop",
		},
		// target wrong domain
		{
			"source":       "http://someblog.com",
			"sourceBody":   "<a href='http://myblog.com'>link</a>",
			"target":       "http://myblog.com",
			"responseCode": fmt.Sprintf("%d", 400),
			"responseBody": "Target: http://myblog.com does not match this site URL: http://monkinetic.blog",
		},
		{
			"source":       "http://someblog.com",
			"sourceBody":   "<a href='http://monkinetic.blog/2020/06/20/test'>link</a>",
			"target":       "http://monkinetic.blog/2020/06/20/test",
			"responseCode": fmt.Sprintf("%d", 201),
			"responseBody": "",
		},
	}

	err := initDb(testDb)
	assert.Nil(t, err)
	defer func() {
		// Teardown
		os.Remove(testDb)
	}()

	assert.Nil(t, err)

	logger.Debugf("config file: %v", testConfigDir)
	var CONFIG = LoadConfig(testConfigDir)
	logger.Debugf("config: %v", CONFIG)

	for _, paramSet := range params {
		source := paramSet["source"]
		sourceBody := paramSet["sourceBody"]
		target := paramSet["target"]
		codeStr := paramSet["responseCode"]
		code, _ := strconv.ParseInt(codeStr, 10, 32)
		body := paramSet["responseBody"]

		// sourceUrl, target
		data := fmt.Sprintf(
			"source=%s&target=%s", source, target)

		logger.Debugf("data: %s", data)

		req, err := http.NewRequest("POST", "/webmention", strings.NewReader(data))
		req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		dbs := &MockDBStorage{}
		dbs.SavePost(&Post{Title: "title", Slug: "test"})

		handler := CreateWebMentionFunc(CONFIG, dbs, &MockPostsRepo{}, &MockWebmentionClient{
			mockedBody: sourceBody,
			mockedCode: http.StatusOK,
		})

		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		respBody, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, int(code), rr.Code)
		assert.Equal(t, body, strings.Trim(string(respBody), "\n"))
	}
}

// TODO: TestCreatePostHandlerPost
