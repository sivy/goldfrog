package syndication

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

type Hook interface {
	HandlePost(postData PostData)
}

type Syndicator interface {
	Hook
	LinkForID(id string) string
}

func MakeSyndicators(config SyndicateConfig) map[string]Hook {
	var posters = make(map[string]Hook)

	posters["twitter"] = NewTwitterPoster(config.Twitter)

	posters["mastodon"] = NewMastodonPoster(config.Mastodon)

	posters["webmention"] = NewWebMentionPoster(config.WebMention)

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
