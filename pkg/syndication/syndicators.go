package syndication

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

type Hook interface {
	HandlePost(postData PostData, linkOnly bool)
}

type Syndicator interface {
	Hook
	LinkForID(id string) string
}

func MakeSyndicators(config SyndicateConfig) map[string]Hook {
	var posters = make(map[string]Hook)

	posters["twitter"] = &TwitterPoster{
		Config:       config,
		ClientKey:    config.Twitter.ClientKey,
		ClientSecret: config.Twitter.ClientSecret,
		AccessKey:    config.Twitter.AccessKey,
		AccessSecret: config.Twitter.AccessSecret,
		LinkFormat:   config.Twitter.LinkFormat,
	}
	posters["mastodon"] = &MastodonPoster{
		Config:       config,
		Site:         config.Mastodon.Site,
		ClientID:     config.Mastodon.ClientID,
		ClientSecret: config.Mastodon.ClientSecret,
		AccessToken:  config.Mastodon.AccessToken,
		LinkFormat:   config.Twitter.LinkFormat,
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
