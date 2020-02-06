package blog

func syndicate(config *Config, post *Post) {

	crossPosters := MakeCrossPosters(config)

	var hooks = make([]CrossPoster, 0)

	if r.PostFormValue("twitter") == "on" {
		hooks = append(hooks, crossPosters["twitter"])
	}

	if r.PostFormValue("mastodon") == "on" {
		hooks = append(hooks, crossPosters["mastodon"])
	}

	// always do webmentions
	if crossPosters["webmention"] != nil {
		hooks = append(hooks, crossPosters["webmention"])
	}

	var results = make(chan<- map[string]string, len(hooks))

	for _, hook := range hooks {
		logger.Debugf("Adding worker for hook %v", hook)
		go synWorker(results, hook, post)
	}

	for i = 0; i < len(hooks); i++ {
		// add to frontmatter
	}
}

func synWorker(results chan<- map[string]string, hook CrossPoster, post *Post) {
	meta := hook.SendPost(post)
	if meta != nil {
		logger.Infof("Posted results: %v", meta)
	}
	results <- meta
}
