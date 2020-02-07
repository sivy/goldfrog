package syndication

import (
	"fmt"
	"html/template"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

type SyndicateConfig struct {
	Twitter    TwitterOpts
	Mastodon   MastodonOpts
	WebMention WebmentionOpts
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

	// var resultQueue = make(chan map[string]string) // len(hooks)

	var wg sync.WaitGroup
	var meta = make(map[string]string)
	for _, hook := range hooks {
		logger.Debugf("Adding worker for hook %v", hook)
		wg.Add(1)
		// go func(results chan map[string]string) {
		go func(h Hook) {
			defer wg.Done()
			res := h.HandlePost(postData)
			logger.Debugf("worker handlePost result: %v", res)
			for k, v := range res {
				meta[k] = v
			}
			logger.Debugf("meta afte hook %v", meta)
		}(hook)
	}
	// close(resultQueue)
	logger.Debug("Waiting...")
	wg.Wait()

	// for i := 0; i < len(hooks); i++ {
	// 	res := <-resultQueue
	// 	logger.Debugf("result from channel: %v", res)
	// 	for k, v := range res {
	// 		meta[k] = v
	// 	}
	// }
	// for res := range resultQueue {
	// 	logger.Debugf("result from channel: %v", res)
	// 	for k, v := range res {
	// 		meta[k] = v
	// 	}
	// }
	return meta
}

// func worker(results chan map[string]string, hook Hook, postData PostData) {
// 	res := hook.HandlePost(postData)
// 	logger.Debugf("worker handlePost result: %v", res)
// 	results <- res
// }

func markDowner(args ...interface{}) template.HTML {
	extensions := parser.CommonExtensions | parser.HeadingIDs
	parser := parser.NewWithExtensions(extensions)
	content := fmt.Sprintf("%s", args...)
	s := markdown.ToHTML(
		[]byte(content), parser, nil)

	return template.HTML(s)
}
