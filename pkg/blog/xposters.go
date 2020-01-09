package blog

import (
	"context"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
)

type CrossPoster interface {
	SendPost(post Post, linkOnly bool) string
}

type TwitterPoster struct {
	Config       Config
	ClientKey    string
	ClientSecret string
	AccessKey    string
	AccessSecret string
}

func (tp *TwitterPoster) SendPost(post Post, linkOnly bool) string {
	config := oauth1.NewConfig(
		tp.Config.Twitter.ClientKey,
		tp.Config.Twitter.ClientSecret)
	token := oauth1.NewToken(
		tp.Config.Twitter.AccessKey,
		tp.Config.Twitter.AccessSecret,
	)

	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	var content string

	if !linkOnly {
		content = post.Body
	}

	content = makeMicroMessage(
		content, 280, post.Title, tp.Config.Blog.Url+post.Url())

	tweet, _, err := client.Statuses.Update(
		content, &twitter.StatusUpdateParams{})

	if err != nil {
		fmt.Printf("%v", err)
		return ""
	}

	log.Infof("%v", tweet)

	return fmt.Sprintf(
		"https://twitter.com/%s/status/%s",
		tweet.User.ScreenName,
		tweet.IDStr)
}

type MastodonPoster struct {
	Config       Config
	Site         string
	ClientID     string
	ClientSecret string
}

func (xp *MastodonPoster) SendPost(post Post, linkOnly bool) string {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       xp.Site,
		ClientID:     xp.ClientID,
		ClientSecret: xp.ClientSecret,
	})
	err := c.Authenticate(context.Background(), "your-email", "your-password")
	if err != nil {
		log.Fatal(err)
	}

	var content string
	if !linkOnly {
		content = post.Body
	}

	content = makeMicroMessage(
		content, 400, post.Title, xp.Config.Blog.Url+post.Url())

	toot := mastodon.Toot{
		Status: content,
	}
	timeline, err := c.PostStatus(context.Background(), toot)

	return ""
}

func MakeCrossPosters(config Config) map[string]CrossPoster {
	var posters = make(map[string]CrossPoster)

	posters["twitter"] = &TwitterPoster{
		Config:       config,
		ClientKey:    config.Twitter.ClientKey,
		ClientSecret: config.Twitter.ClientSecret,
		AccessKey:    config.Twitter.AccessKey,
		AccessSecret: config.Twitter.AccessSecret,
	}
	posters["mastodon"] = &MastodonPoster{
		Config:       config,
		Site:         config.Mastodon.Site,
		ClientKey:    config.Mastodon.ClientKey,
		ClientSecret: config.Mastodon.ClientSecret,
	}
	return posters
}

func statusMessageFromPost(post Post, maxLength int) string {
	body := post.Body

	plaintext := stripHTML(body)

	if len(body) < maxLength {
		maxLength = len(body)
	}

	return plaintext[:maxLength]
}

func stripHTML(args ...interface{}) string {
	content := fmt.Sprintf("%s", args...)
	extensions := parser.CommonExtensions
	parser := parser.NewWithExtensions(extensions)

	s := markdown.ToHTML([]byte(content), parser, nil)

	p := bluemonday.StrictPolicy()
	plaintext := p.Sanitize(string(s))

	return plaintext
}
