package blog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (wc *MockWebmentionClient) FindLinks(htmlStr string) ([]string, error) {
	return wc.mockedLinks, nil
}

type WebMentionTestOpts struct {
	source       string
	sourceBody   string
	target       string
	responseCode int
	responseBody string
}

func TestWebMentionHandler(t *testing.T) {
	// Setup
	var params = []WebMentionTestOpts{
		// bad source and target the same
		{
			"",
			"",
			"",
			400,
			"Webmention source:  cannot be the same as the target: ",
		},
		{
			"http://someblog.com",
			"",
			"http://someblog.com",
			400,
			"Webmention source: http://someblog.com cannot be the same as the target: http://someblog.com",
		},
		// unparseable source
		{
			"floop",
			"",
			"",
			400,
			"Could not parse source: floop",
		},
		// unparseable scheme
		{
			"gopher://example.com",
			"",
			"",
			400,
			"Webmention source must be http(s), not gopher",
		},
		// unparseable target
		{
			"http://someblog.com",
			"",
			"floop",
			400,
			"Could not parse target: floop",
		},
		// unparseable target scheme
		{
			"http://someblog.com",
			"",
			"gopher://example.com",
			400,
			"Webmention target must be http(s), not gopher",
		},
		// target wrong domain
		{
			"http://someblog.com",
			`<a href='http://myblog.com'>link</a>`,
			"http://myblog.com",
			400,
			"Target: http://myblog.com does not match this site URL: http://monkinetic.blog",
		},
		{
			"http://someblog.com",
			`<a href="http://example.com">random link</a> <a href='http://monkinetic.blog/2020/06/20/test'>link</a>`,
			"http://monkinetic.blog/2020/06/20/test",
			201,
			"",
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

	for _, opts := range params {
		source := opts.source
		sourceBody := opts.sourceBody
		target := opts.target
		code := opts.responseCode
		body := opts.responseBody

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
