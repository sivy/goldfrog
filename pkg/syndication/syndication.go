package syndication

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

type SyndicateConfig struct {
}

type PostData struct {
	Title    string
	Slug     string
	PostDate time.Time
	Tags     []string
	Body     string
}

func Syndicate(config SyndicateConfig, includeSyndicators map[string]bool, postData PostData) map[string]string {

	syndicators := MakeSyndicators(config)

	var hooks = make([]Hook, 0)

	for _, label := range []string{"twitter", "mastodon"} {
		if _, ok := includeSyndicators[label]; ok {
			if includeSyndicators[label] {
				hooks = append(hooks, syndicators[label])
			}
		}
	}

	// always do webmentions
	if syndicators["webmention"] != nil {
		hooks = append(hooks, syndicators["webmention"])
	}

	var wg sync.WaitGroup
	for _, hook := range hooks {
		logger.Debugf("Adding worker for hook %v", hook)
		wg.Add(1)
		go worker(hook, postData, &wg)
	}
	logger.Debug("Waiting...")
	wg.Wait()

}

func worker(hook Hook, postData PostData, wg *sync.WaitGroup) {
	defer wg.Done()
	hook.HandlePost(postData)
}

func markDowner(args ...interface{}) template.HTML {
	extensions := parser.CommonExtensions | parser.HeadingIDs
	parser := parser.NewWithExtensions(extensions)
	content := fmt.Sprintf("%s", args...)
	s := markdown.ToHTML(
		[]byte(content), parser, nil)

	return template.HTML(s)
}
