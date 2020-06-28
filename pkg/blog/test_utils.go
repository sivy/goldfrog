package blog

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sivy/goldfrog/pkg/webmention"
)

const (
	testDb        string = "../../tests/data/test.db"
	testConfigDir string = "../../tests/data/"
)

/*
Mock file repo
*/
type MockPostsRepo struct {
	mockFiles []string
}

func (mpr *MockPostsRepo) ListPostFiles() []string {
	return mpr.mockFiles
}

func (mpr *MockPostsRepo) SavePostFile(post *Post) error {
	return nil
}

func (mpr *MockPostsRepo) DeletePostFile(post *Post) error {
	return nil
}

/*
Mock webmention client
*/
type MockWebmentionClient struct {
	mockedBody     string
	mockedCode     int
	mockedEndpoint string
	mockedLinks    []string
}

func (wc *MockWebmentionClient) Fetch(url string) (*http.Response, error) {
	logger.Debugf("mock fetch for %s", url)
	bodyIo := strings.NewReader(wc.mockedBody)
	resp := http.Response{
		Body:       ioutil.NopCloser(bodyIo),
		StatusCode: wc.mockedCode,
		Request:    &http.Request{},
	}
	return &resp, nil
}

func (wc *MockWebmentionClient) EndpointDiscovery(mentionTarget string) (string, error) {
	return wc.mockedEndpoint, nil
}

func (wc *MockWebmentionClient) SendWebMentions(source string, links []string) {
	// NOOP
}

func (wc *MockWebmentionClient) SendMention(endpoint string, source string, target string) {
	// NOOP
}
func (wc *MockWebmentionClient) GetHtmlEndpoint(doc *goquery.Document, elements []string) string {
	return wc.mockedEndpoint
}
func (wc *MockWebmentionClient) GetMention(targetUrl string, r io.Reader) (webmention.WebMention, error) {
	return webmention.WebMention{}, nil
}

type MockDBStorage struct {
	mockPosts          map[string]*Post
	mockPost           *Post
	mockArchiveEntries []ArchiveEntry
}

func (dbs *MockDBStorage) GetPosts(opts GetPostOpts) []*Post {
	var v = make([]*Post, 0)
	for _, value := range dbs.mockPosts {
		v = append(v, value)
	}
	return v
}
func (dbs *MockDBStorage) GetTaggedPosts(tag string) []*Post {
	return dbs.GetPosts(GetPostOpts{})
}
func (dbs *MockDBStorage) GetPost(postID string) (*Post, error) {
	return dbs.mockPost, nil
}
func (dbs *MockDBStorage) GetPostBySlug(postSlug string) (*Post, error) {
	return dbs.mockPosts[postSlug], nil
}
func (dbs *MockDBStorage) GetArchiveYearMonths() []ArchiveEntry {
	return dbs.mockArchiveEntries
}
func (dbs *MockDBStorage) GetArchiveMonthPosts(year string, month string) []*Post {
	return dbs.GetPosts(GetPostOpts{})
}
func (dbs *MockDBStorage) GetArchiveDayPosts(year string, month string, day string) []*Post {
	return dbs.GetPosts(GetPostOpts{})
}
func (dbs *MockDBStorage) CreatePost(post *Post) error {
	post.ID = 1
	dbs.mockPost = post
	if dbs.mockPosts == nil {
		dbs.mockPosts = make(map[string]*Post, 0)
	}
	dbs.mockPosts[post.Slug] = post
	return nil
}
func (dbs *MockDBStorage) SavePost(post *Post) error {
	dbs.mockPost = post
	if dbs.mockPosts == nil {
		dbs.mockPosts = make(map[string]*Post, 0)
	}
	dbs.mockPosts[post.Slug] = post
	return nil
}
func (dbs *MockDBStorage) DeletePost(postID string) error {
	dbs.mockPost = nil
	dbs.mockPosts = nil
	return nil
}

func NewMockDBStorage() DBStorage {
	return &MockDBStorage{}
}
