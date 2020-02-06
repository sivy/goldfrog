package syndication

import (
	"context"
	"fmt"

	mastodon "github.com/mattn/go-mastodon"
)

const (
	mastodonMaxMessageLen int = 500
)

type MastodonPoster struct {
	Site         string
	ClientID     string
	ClientSecret string
	AccessToken  string
	LinkFormat   string
}

func (xp *MastodonPoster) HandlePost(postData *PostData) {
	logger.Infof("Handling Mastodon crosspost...")
	c := mastodon.NewClient(&mastodon.Config{
		Server:       xp.Site,
		ClientID:     xp.ClientID,
		ClientSecret: xp.ClientSecret,
		AccessToken:  xp.AccessToken,
	})

	var content string
	if !linkOnly {
		content = post.Body
	}
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
	}

	post.FrontMatter["mastodon_id"] = string(status.ID)
	post.FrontMatter["mastodon_url"] = status.URL

	logger.Debugf("Posted status: %s", status.URL)
}

func (xp *MastodonPoster) LinkForID(config Config, id string) string {
	logger.Debugf("link format: %s", config.Mastodon.LinkFormat)
	return fmt.Sprintf(config.Mastodon.LinkFormat, id)
}
