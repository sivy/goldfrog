/*
3.2 Receiving Webmentions
*/
package blog

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sivy/goldfrog/pkg/webmention"
)

func CreateWebMentionFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating webmention handler")

	//	author_tz, _ := time.LoadLocation(config.Blog.Author.TimeZone)

	return func(w http.ResponseWriter, r *http.Request) {

		logger.Info("Handling webmention")

		sourceUrl := r.PostFormValue("source")
		target := r.PostFormValue("target")

		/*
			3.2.1 Request Verification

			The receiver MUST reject the request if the source URL is the same as the
			target URL.
		*/
		if sourceUrl == target {
			message := fmt.Sprintf(
				"Webmention source: %s cannot be the same as the target: %s",
				sourceUrl, target)
			logger.Error(message)
			http.Error(w, message, 400)
			return
		}

		/*
			3.2.1 Request Verification

			The receiver MUST check that source and target are valid URLs [URL] and are
			of schemes that are supported by the receiver. (Most commonly this means
			checking that the source and target schemes are http or https).
		*/
		parsedSource, err := url.ParseRequestURI(sourceUrl)
		if err != nil {
			logger.Errorf("Could not parse source: %s, %s", sourceUrl, err)
			http.Error(w, err.Error(), 400)
			return
		}
		if !(parsedSource.Scheme == "http") && !(parsedSource.Scheme == "https") {
			message := "Webmention source must be http(s)"
			logger.Error(message)
			http.Error(w, message, 400)
			return
		}

		parsedTarget, err := url.ParseRequestURI(target)
		if err != nil {
			logger.Errorf("Could not parse target: %s, %s", target, err)
			http.Error(w, err.Error(), 400)
			return
		}
		if !(parsedTarget.Scheme == "http") && !(parsedTarget.Scheme == "https") {
			message := "Webmention target must be http(s)"
			logger.Error(message)
			http.Error(w, message, 400)
			return
		}

		if !(strings.Contains(target, config.Blog.Url)) {
			message := fmt.Sprintf(
				"Target: %s does not match this site URL: %s", target, config.Blog.Url)
			logger.Error(message)
			http.Error(w, message, 400)
			return
		}

		pathBits := strings.Split(parsedTarget.Path, "/")
		targetSlug := pathBits[len(pathBits)-1]

		/*
			3.2.1 Request Verification

			The receiver SHOULD check that target is a valid resource for which it can
			accept Webmentions. This check SHOULD happen synchronously to reject invalid
			Webmentions before more in-depth verification begins.
		*/
		post, err := GetPostBySlug(db, targetSlug)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 404)
			return
		}

		/*
			3.2.2 Webmention Verification

			TODO: Webmention verification SHOULD be handled asynchronously to prevent DoS (Denial of Service) attacks.
		*/
		logger.Debugf(
			"Webmention verification from: %s to: %s (post %s)",
			sourceUrl, target, post.Title)

		/*
			If the receiver is going to use the Webmention in some way, (displaying it as
			a comment on a post, incrementing a "like" counter, notifying the author of a
			post), then it MUST perform an HTTP GET request on source, following any HTTP
			redirects (and SHOULD limit the number of redirects it follows) to confirm
			that it actually mentions the target
		*/
		client := webmention.Client{}
		sourceResp, err := client.Fetch(sourceUrl)

		if err != nil {
			logger.Errorf("Could not load source: %s, %s", sourceUrl, err)
			return
		}

		doc, err := goquery.NewDocumentFromResponse(sourceResp)
		if err != nil {
			logger.Errorf("Could not load source: %s, %s", sourceUrl, err)
			http.Error(w, err.Error(), 500)
			return
		}

		selectors := []string{
			"a",
		}

		selector := strings.Join(selectors, ",")
		var foundLink bool

		doc.Find(selector).EachWithBreak(
			func(i int, s *goquery.Selection) bool {
				_, ok := s.Attr("href")
				if ok {
					foundLink = ok
					return false
				}
				return true
			})

		if !foundLink {
			message := fmt.Sprintf("Source: %s does not link to target: %s", sourceUrl, target)
			logger.Errorf(message)
			http.Error(w, message, 400)
			return
		}

		if _, ok := post.FrontMatter["webmentions"]; ok {
			//do something here
		}
		// err = repo.SavePostFile(post)
		// if err != nil {
		// 	logger.Errorf("Could not save post file: %v", err)
		// }

		// err = SavePost(db, post)
		// if err != nil {
		// 	logger.Errorf("Could not save post: %v", err)
		// }

		// updatePost, err := GetPostBySlug(db, post.Slug)

		// if err != nil {
		// 	logger.Errorf("Post saved but syndication process could not run: %v", err)
		// 	SetFlash(w, "flash", fmt.Sprintf("Post saved but syndication process could not run: %v", err))
		// 	http.Redirect(w, r, "/", http.StatusFound)
		// }

		// if r.PostFormValue("twitter") == "on" {
		// 	includeHooks["twitter"] = true
		// 	synOpts.Twitter = config.Twitter
		// }

		// if r.PostFormValue("mastodon") == "on" {
		// 	includeHooks["mastodon"] = true
		// 	synOpts.Mastodon = config.Mastodon
		// }

		// if config.WebMentionEnabled {
		// 	includeHooks["webmention"] = true
		// 	synOpts.WebMention = syndication.WebmentionOpts{}
		// }

		// don't depend on updating a reference to a Post
		// postData := syndication.PostData{
		// 	Title:       updatePost.Title,
		// 	Slug:        updatePost.Slug,
		// 	Tags:        updatePost.Tags,
		// 	Body:        updatePost.Body,
		// 	FrontMatter: updatePost.FrontMatter,
		// 	PermaLink:   config.Blog.Url + updatePost.PermaLink(),
		// }
		// if !postDate.IsZero() {
		// 	postData.PostDate = postDate
		// }

		// syndicationMeta := syndication.Syndicate(synOpts, includeHooks, postData)

		// logger.Debugf("new meta after hooks: %v", syndicationMeta)

		// for k, v := range syndicationMeta {
		// 	updatePost.FrontMatter[k] = v
		// }

		// err = SavePost(db, updatePost)
		// if err != nil {
		// 	logger.Error(err)
		// 	SetFlash(w, "flash", fmt.Sprintf(
		// 		"Your post was saved, but some syndication links might be missing (%v)",
		// 		err))
		// }

		// err = repo.SavePostFile(updatePost)
		// if err != nil {
		// 	logger.Error(err)
		// 	SetFlash(w, "flash", fmt.Sprintf(
		// 		"Your post was saved, but some syndication links might be missing on disk (%v)",
		// 		err))
		// }

		// redirect(w, config.TemplatesDir, post.PermaLink())
		// http.Redirect(w, r, post.PermaLink(), http.StatusFound)
		return
	}
}
