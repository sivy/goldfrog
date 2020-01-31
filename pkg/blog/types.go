package blog

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
	"time"
)

type Blog struct {
	Title   string            `json:"title"`
	Subhead string            `json:"subhead"`
	Author  string            `json:"author"`
	Meta    map[string]string `json:"meta"`
}

type PostOpts struct {
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	PostDate    time.Time         `json:"date"`
	Tags        []string          `json:"tags"`
	FrontMatter map[string]string `json:"frontmatter"`
	Body        string            `json:"body"`
}

type Post struct {
	ID          int               `json:"post_id"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	PostDate    time.Time         `json:"date"`
	Tags        []string          `json:"tags"`
	FrontMatter map[string]string `json:"frontmatter"`
	Body        string            `json:"body"`
	User        User              `json:"user"`
}

func (post *Post) TagString() string {
	return strings.Join(post.Tags, ", ")
}

func (post *Post) PermaLink() string {
	return fmt.Sprintf(
		"/%s/%s",
		post.PostDate.Format("2006/01/02"),
		post.Slug,
	)
}

func (post *Post) PermaShortId() string {
	return post.Slug
}

func (post *Post) ToString() string {
	fm := post.FrontMatter
	fm["title"] = post.Title
	fm["slug"] = post.Slug
	fm["date"] = post.PostDate.Format(POSTTIMESTAMPFMT)
	fm["tags"] = strings.Join(post.Tags, ",")

	fmBytes, _ := yaml.Marshal(fm)

	fileContent := string(fmBytes)
	fileContent = fileContent + fmt.Sprintf("---\n")
	fileContent = fileContent + fmt.Sprintf("%s\n", post.Body)

	return fileContent
}

func NewPost(opts PostOpts) Post {
	p := Post{
		Title:    opts.Title,
		Slug:     opts.Slug,
		PostDate: opts.PostDate,
		Tags:     opts.Tags,
		Body:     opts.Body,
	}
	if opts.FrontMatter != nil {
		p.FrontMatter = opts.FrontMatter
	} else {
		p.FrontMatter = make(map[string]string)
	}
	return p
}

type User struct {
	DisplayName string `json:"displayname"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Url         string `json:"url"`
	Image       string `json:"image"`
	IsAdmin     bool   `json:"isadmin"`
}
