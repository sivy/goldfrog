package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sivy/goldfrog/pkg/blog"
)

type Tweet struct {
	FullText  string `json:"full_text"`
	IDStr     string `json:"id_str"`
	CreatedAt string `json:"created_at"`
}

type TweetArchive struct {
	Tweet Tweet `json:"tweet"`
}

func loadTweetFile(path string) ([]TweetArchive, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	var tweetData []TweetArchive
	err = json.Unmarshal(data, &tweetData)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	return tweetData, nil
}

func tweetToPost(ta TweetArchive) (blog.Post, error) {
	var body string
	var date time.Time
	var tags []string
	var frontMatter = make(map[string]string)

	body = ta.Tweet.FullText
	fmt.Println(fmt.Sprintf("tweet created_at: %v", ta.Tweet.CreatedAt))
	if ta.Tweet.CreatedAt != "" {
		layout, err := dateparse.ParseFormat(ta.Tweet.CreatedAt)
		if err != nil {
			fmt.Println(fmt.Sprintf("format error! %v", err))
			return blog.NewPost(blog.PostOpts{}), err
		}

		parsedDate, err := time.Parse(layout, ta.Tweet.CreatedAt)
		if err != nil {
			fmt.Println(fmt.Sprintf("parse error! %v", err))
			return blog.NewPost(blog.PostOpts{}), err
		}
		fmt.Println(fmt.Sprintf("parsed date: %v", parsedDate))

		date = parsedDate
	}
	fmt.Println(fmt.Sprintf("using date: %v", date))
	tags = blog.GetHashTags(body)

	frontMatter["twitter_id"] = ta.Tweet.IDStr
	frontMatter["twitter_url"] = fmt.Sprintf(
		"https://twitter.com/steveivy/statuses/%s", ta.Tweet.IDStr)

	post := blog.NewPost(blog.PostOpts{
		Body:        body,
		Tags:        tags,
		PostDate:    date,
		FrontMatter: frontMatter,
		Slug:        blog.MakeNoteSlug(body),
	})

	return post, nil
}

func main() {

	var tweetsFile string
	var tweetJson []TweetArchive
	var postsDir string

	flag.StringVar(
		&tweetsFile, "tweetfile",
		"tweets.json",
		"Location of your posts (Jekyll-compatible markdown)")

	flag.StringVar(
		&postsDir, "posts_dir",
		"",
		"Location of your posts (Jekyll-compatible markdown)")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Loading %s", tweetsFile))

	if tweetsFile != "" {
		parsedJson, err := loadTweetFile(tweetsFile)
		if err == nil {
			fmt.Printf("Found %d tweets", len(tweetJson))
		}
		tweetJson = parsedJson
	}

	repo := blog.FilePostsRepo{
		PostsDirectory: postsDir,
	}

	for _, ta := range tweetJson {
		post, _ := tweetToPost(ta)
		fmt.Println(fmt.Sprintf("%v: %s", post.PostDate, post.Body))
		if postsDir != "" {
			repo.SavePostFile(&post)
		}
	}

}
