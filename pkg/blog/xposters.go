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
	"github.com/sivy/goldfrog/pkg/webmention"
)

type CrossPoster interface {
	SendPost(post *Post, linkOnly bool) map[string]string
}

type TwitterPoster struct {
	Config       Config
	ClientKey    string
	ClientSecret string
	AccessKey    string
	AccessSecret string
}

func (tp *TwitterPoster) SendPost(post *Post) map[string]string {
	logger.Infof("Handling Twitter crosspost...")
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

	var content = post.Body

	opts := MicroMessageOpts{
		MaxLength: 280,
	}

	if post.Title != "" {
		opts.Title = post.Title
		opts.PermaLink = tp.Config.Blog.Url + post.PermaLink()
	} else {
		opts.ShortID = post.Slug
	}

	content = makeMicroMessage(content, opts)

	tweet, _, err := client.Statuses.Update(
		content, &twitter.StatusUpdateParams{})

	if err != nil {
		fmt.Printf("%v", err)
		return ""
	}

	url := fmt.Sprintf(
		"https://twitter.com/%s/status/%s",
		tweet.User.ScreenName,
		tweet.IDStr)

	logger.Debugf("Posted status: %s", url)
	return map[string]string{
		"twitter_id":  tweet.IDStr,
		"twitter_url": url,
	}
}

type MastodonPoster struct {
	Config       Config
	Site         string
	ClientID     string
	ClientSecret string
	AccessToken  string
}

func (xp *MastodonPoster) SendPost(post *Post) map[string]string {
	logger.Infof("Handling Mastodon crosspost...")
	c := mastodon.NewClient(&mastodon.Config{
		Server:       xp.Site,
		ClientID:     xp.ClientID,
		ClientSecret: xp.ClientSecret,
		AccessToken:  xp.AccessToken,
	})

	var content = post.Body

	opts := MicroMessageOpts{
		MaxLength: 280,
	}

	if post.Title != "" {
		opts.Title = post.Title
		opts.PermaLink = xp.Config.Blog.Url + post.PermaLink()
	} else {
		opts.ShortID = post.Slug
	}

	content = makeMicroMessage(content, opts)

	toot := mastodon.Toot{
		Status:     content,
		Visibility: "unlisted",
	}

	if post.Title != "" {
		toot.SpoilerText = post.Title
		toot.Sensitive = true
	}

	logger.Debugf("Sending Mastodon post...")
	status, err := c.PostStatus(context.Background(), &toot)
	if err != nil {
		logger.Error(err)
		return ""
	}
	logger.Debugf("Posted status: %s", status.URL)
	return map[string]string{
		"mastodon_id":  status.ID,
		"mastodon_url": status.URL,
	}
}

type WebMentionPoster struct {
	Config Config
}

func (wp *WebMentionPoster) SendPost(post *Post, linkOnly bool) map[string]string {
	logger.Infof("Handling WebMentions...")
	client := webmention.NewWebMentionClient()
	htmlText := string(markDowner(post.Body))

	sourceLink := wp.Config.Blog.Url + post.PermaLink()
	links, err := client.FindLinks(htmlText)
	if err != nil {
		logger.Errorf("Could not get post links: %s", err)
		return ""
	}
	logger.Debugf("Found links: %v", links)
	logger.Info("Sending WebMentions...")
	client.SendWebMentions(sourceLink, links)

	return make(map[string]string)
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
		ClientID:     config.Mastodon.ClientID,
		ClientSecret: config.Mastodon.ClientSecret,
		AccessToken:  config.Mastodon.AccessToken,
	}
	if config.WebMentionEnabled {
		posters["webmention"] = &WebMentionPoster{
			Config: config,
		}
	} else {
		posters["webmention"] = nil
	}
	return posters
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
