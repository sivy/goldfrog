package blog

import (
	"fmt"
	"strings"
	"time"
)

type Blog struct {
	Title   string            `json:"title"`
	Subhead string            `json:"subhead"`
	Author  string            `json:"author"`
	Meta    map[string]string `json:"meta"`
}

type Post struct {
	ID       int       `json:"post_id"`
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Tags     []string  `json:"tags"`
	PostDate time.Time `json:"date"`
	Body     string    `json:"body"`
	Format   string    `json:"format"`
}

func (post *Post) TagString() string {
	return strings.Join(post.Tags, ", ")
}

func (post *Post) Url() string {
	return fmt.Sprintf(
		"/%s/%s",
		post.PostDate.Format("2006/01"),
		post.Slug,
	)
}

func (post *Post) ToString() string {
	filecontent := fmt.Sprintf("title: %s\n", post.Title)
	filecontent = filecontent + fmt.Sprintf("slug: %s\n", post.Slug)
	filecontent = filecontent + fmt.Sprintf(
		"date: %s\n", post.PostDate.Format("2006-01-02 03:04:05"))
	filecontent = filecontent + fmt.Sprintf("tags: %s\n", strings.Join(post.Tags, ","))
	filecontent = filecontent + fmt.Sprintf("---\n")
	filecontent = filecontent + fmt.Sprintf("%s\n", post.Body)

	return filecontent
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}
