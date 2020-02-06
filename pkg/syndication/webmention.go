package syndication

import (
	"github.com/sivy/goldfrog/pkg/webmention"
)

type WebMentionPoster struct {
}

func (wp *WebMentionPoster) HandlePost(postData PostData) {
	logger.Infof("Handling WebMentions...")
	client := webmention.NewWebMentionClient()
	htmlText := string(markDowner(postData.Body))

	sourceLink := postData.PermaLink
	links, err := client.FindLinks(htmlText)
	if err != nil {
		logger.Errorf("Could not get post links: %s", err)
	}
	logger.Debugf("Found links: %v", links)
	logger.Info("Sending WebMentions...")
	client.SendWebMentions(sourceLink, links)
}

func NewWebMentionPoster(WebmentionOpts) *WebMentionPoster {
	return &WebMentionPoster{}
}
