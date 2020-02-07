package syndication

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	MaxLen       int
}

func (xp *MastodonPoster) FormatMessage(postData PostData) string {

	source := stripHTML(markDowner(postData.Body))
	var title string
	var link string

	if postData.Title != "" {
		title = postData.Title
	}

	if postData.PermaLink != "" {
		link = fmt.Sprintf("(%s)", postData.PermaLink)
	}
	if postData.ShortID != "" {
		link = fmt.Sprintf("(monkinetic %s)", postData.ShortID)
	}

	var fmtTagStr string
	if len(postData.Tags) != 0 {
		var fmtTags []string
		for _, t := range postData.Tags {
			if t == "" {
				continue
			}
			if regexp.MustCompile("#" + t).MatchString(source) {
				continue
			}
			fmtTags = append(fmtTags, fmt.Sprintf("#%s", t))
		}
		fmtTagStr = strings.Join(fmtTags, " ")
	}

	var messageParts []string
	availableChars := xp.MaxLen
	if title != "" {
		availableChars -= len(title) + 2 // len(\n\n)
	}
	if link != "" {
		availableChars -= len(link) + 2 // len(\n\n)
	}
	if fmtTagStr != "" {
		availableChars -= len(fmtTagStr) + 2 // len(\n\n)
	}

	// split paras
	sourceParas := strings.Split(source, "\n\n")
	var messageBody string

	if len(source) < availableChars {
		// if message fits, do it all
		messageBody = source
	} else {
		// if it doesn't, only do first para
		messageBody = sourceParas[0]
	}

	// find closes para that fits in available length

	if title != "" {
		messageParts = append(messageParts, title)
	}
	messageParts = append(messageParts, messageBody)

	if link != "" {
		messageParts = append(messageParts, link)
	}
	if fmtTagStr != "" {
		messageParts = append(messageParts, fmtTagStr)
	}

	microMessage := strings.Join(messageParts, "\n\n")
	logger.Debugf("microMessage: %s", microMessage)
	return microMessage
}

func (xp *MastodonPoster) HandlePost(postData PostData) {
	logger.Infof("Handling Mastodon crosspost...")
	c := mastodon.NewClient(&mastodon.Config{
		Server:       xp.Site,
		ClientID:     xp.ClientID,
		ClientSecret: xp.ClientSecret,
		AccessToken:  xp.AccessToken,
	})

	if postData.Title == "" {
		postData.ShortID = postData.Slug
	}

	var content = xp.FormatMessage(postData)

	toot := mastodon.Toot{
		Status:     content,
		Visibility: "unlisted",
	}

	if postData.Title != "" {
		toot.SpoilerText = postData.Title
		toot.Sensitive = true
	}

	logger.Debugf("Sending Mastodon post..")
	status, err := c.PostStatus(context.Background(), &toot)
	if err != nil {
		logger.Error(err)
	}

	postData.FrontMatter["mastodon_id"] = string(status.ID)
	postData.FrontMatter["mastodon_url"] = status.URL

	logger.Debugf("Posted status: %s", status.URL)
}

func (xp *MastodonPoster) LinkForID(id string) string {
	return fmt.Sprintf("%s/status/%s", xp.Site, id)
}

func NewMastodonPoster(opts MastodonOpts) *MastodonPoster {
	return &MastodonPoster{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		AccessToken:  opts.AccessToken,
		MaxLen:       500,
	}
}