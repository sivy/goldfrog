package blog

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/go-chi/chi"
	"github.com/sivy/goldfrog/pkg/syndication"
)

func CreateNewPostFunc(
	config Config, dbs DBStorage, repo PostsRepo) http.HandlerFunc {

	author_tz, _ := time.LoadLocation(config.Blog.Author.TimeZone)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if !checkIsOwner(config, r) {
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			logger.Info("Rendering New Post form")
			t, err := getTemplate(config.TemplatesDir, "newpost.html")
			if err != nil {
				logger.Errorf("Could not get template: %v", err)
			}

			title := r.FormValue("title")
			body := r.FormValue("body")
			slug := r.FormValue("slug")
			tagStr := r.FormValue("tags")
			tags := splitTags(tagStr)

			flash, _ := GetFlash(w, r, "flash")

			err = t.ExecuteTemplate(w, "base", struct {
				Config             Config
				Post               Post
				PostDateInTimeZone time.Time
				FormAction         string
				IsOwner            bool
				TextHeight         int
				ShowSlug           bool
				Flash              string
			}{
				Config: config,
				Post: NewPost(PostOpts{
					Title: title,
					Body:  body,
					Slug:  slug,
					Tags:  tags,
				}),
				PostDateInTimeZone: time.Now().In(author_tz),
				FormAction:         "/new",
				IsOwner:            true,
				TextHeight:         20,
				ShowSlug:           true,
				Flash:              flash,
			})
			return
		}

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("postimage")
		var hasImage bool
		var imageUrl string
		var imagePath string
		var mediaBytes []byte
		var mediaType string

		if err == nil {
			// there's an image
			hasImage = true
			defer file.Close()
			logger.Infof("File upload in progress...")

			imagePath = filepath.Join(config.UploadsDir, handler.Filename)
			logger.Debugf("Writing uploaded file: %s", imagePath)
			f, err := os.OpenFile(
				imagePath, os.O_WRONLY|os.O_CREATE, 0777)

			if err != nil {
				logger.Error(err)
				hasImage = false
			} else {
				defer f.Close()
				io.Copy(f, file)
				imageUrl = strings.Join([]string{
					config.Blog.Url, "uploads", handler.Filename}, "/")
			}

			mediaBytes, err = ioutil.ReadFile(imagePath)
			if err != nil {
				logger.Error(err)
				hasImage = false
			} else {
				ext := filepath.Ext(imagePath)
				logger.Debug(mediaType)
				mediaType = "tweet_image"
				if ext == ".gif" {
					mediaType = "tweet_gif"
				}

				logger.Debug(mediaType)
			}
			// get bytes
			// file
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")
		slug := r.PostFormValue("slug")
		date := r.PostFormValue("postdate")

		if title == "" {
			slug = MakeNoteSlug(body)
		}

		if slug == "" {
			slug = MakePostSlug(title)
		}

		body = strings.Replace(body, "\r\n", "\n", -1)

		if hasImage {
			body = strings.Replace(
				body, "[image]",
				fmt.Sprintf("![%s](%s)", imageUrl, imageUrl),
				-1)
		}

		var postDate time.Time
		if date != "" {
			// the date is provided in the author's timezone
			postDate, err = dateparse.ParseIn(date, author_tz)
			if err != nil {
				logger.Warnf("Could not parse date: %s", date)
				date = ""
			}
			// then convert to UTC for storage
			postDate = postDate.UTC()
		}
		if date == "" {
			// server is already in UTC
			postDate = time.Now()
		}

		post := NewPost(PostOpts{
			Title:    title,
			Tags:     splitTags(tags),
			Body:     body,
			Slug:     slug,
			PostDate: postDate,
		})

		logger.Debug(post)
		post.Tags = updateTags(post.Body, post.Tags)

		err = repo.SavePostFile(&post)
		if err != nil {
			logger.Errorf("Could not save post: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Could not save post: %v", err))

			values := url.Values{
				"title": []string{title},
				"slug":  []string{slug},
				"body":  []string{body},
				"tags":  []string{tags},
			}
			http.Redirect(
				w, r,
				fmt.Sprintf("/new?%s", values.Encode()),
				http.StatusSeeOther)
			// redirect(w, config.TemplatesDir, "/")
			return
		}

		err = dbs.CreatePost(&post)
		if err != nil {
			logger.Errorf("Could not create post file: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Could not save post file: %v", err))
			values := url.Values{
				"title": []string{title},
				"slug":  []string{slug},
				"body":  []string{body},
				"tags":  []string{tags},
			}
			http.Redirect(
				w, r,
				fmt.Sprintf("/new?%s", values.Encode()),
				http.StatusSeeOther)
			// redirect(w, config.TemplatesDir, "/")
			return
		}

		updatePost, err := dbs.GetPostBySlug(post.Slug)

		if err != nil {
			logger.Errorf("Post saved but syndication process could not run: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Post saved but syndication process could not run: %v", err))
			http.Redirect(w, r, "/", http.StatusFound)
		}

		includeHooks := make(map[string]bool)
		synOpts := syndication.SyndicateConfig{}

		if r.PostFormValue("twitter") == "on" {
			includeHooks["twitter"] = true
			synOpts.Twitter = config.Twitter
		}

		if r.PostFormValue("mastodon") == "on" {
			includeHooks["mastodon"] = true
			synOpts.Mastodon = config.Mastodon
		}

		if config.WebMentionEnabled {
			includeHooks["webmention"] = true
			synOpts.WebMention = syndication.WebmentionOpts{}
		}

		// don't depend on updating a reference to a Post
		postData := syndication.PostData{
			Title:       updatePost.Title,
			Slug:        updatePost.Slug,
			PostDate:    updatePost.PostDate,
			Tags:        updatePost.Tags,
			Body:        updatePost.Body,
			FrontMatter: updatePost.FrontMatter,
			PermaLink:   config.Blog.Url + updatePost.PermaLink(),
		}
		if len(mediaBytes) > 0 {
			postData.MediaContent = mediaBytes
			postData.MediaType = mediaType
		}
		syndicationMeta := syndication.Syndicate(synOpts, includeHooks, postData)

		logger.Debugf("new meta after hooks: %v", syndicationMeta)

		for k, v := range syndicationMeta {
			updatePost.FrontMatter[k] = v
		}

		err = dbs.SavePost(updatePost)
		if err != nil {
			logger.Error(err)
			SetFlash(w, "flash", fmt.Sprintf(
				"Your post was saved, but some syndication links might be missing (%v)",
				err))
		}

		err = repo.SavePostFile(updatePost)
		if err != nil {
			logger.Error(err)
			SetFlash(w, "flash", fmt.Sprintf(
				"Your post was saved, but some syndication links might be missing on disk (%v)",
				err))
		}

		http.Redirect(w, r, "/", http.StatusFound)
		// redirect(w, config.TemplatesDir, "/")
		return
	}
}

func CreateEditPostFunc(
	config Config, dbs DBStorage, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating edit post handler")

	author_tz, _ := time.LoadLocation(config.Blog.Author.TimeZone)

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			logger.Info("Rendering Edit Post form")

			if !checkIsOwner(config, r) {
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			postID := chi.URLParam(r, "postID")
			logger.Debugf("Post %s", postID)

			t, err := getTemplate(config.TemplatesDir, "editpost.html")
			if err != nil {
				logger.Errorf("Could not get template: %v", err)
			}
			logger.Debug(t)

			post, err := dbs.GetPost(postID)

			if err != nil {
				logger.Error(err)
				flash := fmt.Sprintf("An error occurred: %v", err)

				err = t.ExecuteTemplate(w, "base", struct {
					Config             Config
					Post               *Post
					PostDateInTimeZone time.Time
					FormAction         string
					IsOwner            bool
					TextHeight         int
					ShowSlug           bool
					ShowExpand         bool
					Flash              string
				}{
					Config:             config,
					Post:               post,
					PostDateInTimeZone: post.PostDate.In(author_tz),
					FormAction:         "/edit",
					IsOwner:            true,
					TextHeight:         20,
					ShowSlug:           true,
					ShowExpand:         false,
					Flash:              flash,
				})
			}
			logger.Debugf("Found post %s", post.Title)

			post.User = User{
				DisplayName: config.Blog.Author.Name,
				Email:       config.Blog.Author.Email,
				Url:         config.Blog.Url,
				IsAdmin:     true,
			}

			flash, _ := GetFlash(w, r, "flash")

			err = t.ExecuteTemplate(w, "base", struct {
				Config             Config
				Post               *Post
				PostDateInTimeZone time.Time
				FormAction         string
				IsOwner            bool
				TextHeight         int
				ShowSlug           bool
				ShowExpand         bool
				Flash              string
			}{
				Config:             config,
				Post:               post,
				PostDateInTimeZone: post.PostDate.In(author_tz),
				FormAction:         "/edit",
				IsOwner:            true,
				TextHeight:         20,
				ShowSlug:           true,
				ShowExpand:         false,
				Flash:              flash,
			})
			return
		}
		logger.Info("Handling Edit Post form")

		postID := r.FormValue("postID")
		post, err := dbs.GetPost(postID)
		if err != nil {
			logger.Error(err)
		}

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("postimage")
		var hasImage bool
		var imageUrl string
		var imagePath string

		if err == nil {
			// there's an image
			hasImage = true
			defer file.Close()
			logger.Infof("File upload in progress...")

			imagePath = filepath.Join(config.UploadsDir, handler.Filename)
			logger.Debugf("Writing uploaded file: %s", imagePath)

			f, err := os.OpenFile(
				imagePath, os.O_WRONLY|os.O_CREATE, 0777)
			if err != nil {
				logger.Error(err)
				hasImage = false
			} else {
				defer f.Close()
				io.Copy(f, file)
				imageUrl = strings.Join([]string{
					config.Blog.Url, "uploads", handler.Filename}, "/")
			}
		} else {
			logger.Error(err)
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")
		date := r.PostFormValue("postdate")

		frontMatterString := r.PostFormValue("meta")

		frontMatterYaml := GetFrontMatter(frontMatterString)

		body = strings.Replace(body, "\r\n", "\n", -1)

		if hasImage {
			body = strings.Replace(
				body, "[image]",
				fmt.Sprintf("![%s](%s)", imageUrl, imageUrl),
				-1)
		}

		post.Title = title
		post.Tags = splitTags(tags)
		post.Body = strings.TrimSpace(body)
		post.FrontMatter = frontMatterYaml

		var postDate time.Time
		logger.Infof("Edit post posted date: %v", date)
		if date != "" {
			// the date is provided in the author's timezone
			postDate, err = dateparse.ParseIn(date, author_tz)
			logger.Infof("postDate: %v", postDate)
			if err != nil {
				logger.Warnf("Could not parse date: %s", date)
				date = ""
			}
			// then convert to UTC for storage
			postDate = postDate.UTC()
			logger.Infof("postDate UTC: %v", postDate)
		}
		if date == "" {
			// server is already in UTC
			postDate = time.Now()
		}
		post.PostDate = postDate

		processedBody := fmt.Sprintf("%s", markDowner(post.Body))
		post.Tags = updateTags(processedBody, post.Tags)

		logger.Debug(post)

		err = repo.SavePostFile(post)
		if err != nil {
			logger.Errorf("Could not save post file: %v", err)
		}

		err = dbs.SavePost(post)
		if err != nil {
			logger.Errorf("Could not save post: %v", err)
		}

		updatePost, err := dbs.GetPostBySlug(post.Slug)

		if err != nil {
			logger.Errorf("Post saved but syndication process could not run: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Post saved but syndication process could not run: %v", err))
			http.Redirect(w, r, "/", http.StatusFound)
		}

		includeHooks := make(map[string]bool)
		synOpts := syndication.SyndicateConfig{}

		if r.PostFormValue("twitter") == "on" {
			includeHooks["twitter"] = true
			synOpts.Twitter = config.Twitter
		}

		if r.PostFormValue("mastodon") == "on" {
			includeHooks["mastodon"] = true
			synOpts.Mastodon = config.Mastodon
		}

		if config.WebMentionEnabled {
			includeHooks["webmention"] = true
			synOpts.WebMention = syndication.WebmentionOpts{}
		}

		// don't depend on updating a reference to a Post
		postData := syndication.PostData{
			Title:       updatePost.Title,
			Slug:        updatePost.Slug,
			Tags:        updatePost.Tags,
			Body:        updatePost.Body,
			FrontMatter: updatePost.FrontMatter,
			PermaLink:   config.Blog.Url + updatePost.PermaLink(),
		}
		if !postDate.IsZero() {
			postData.PostDate = postDate
		}

		syndicationMeta := syndication.Syndicate(synOpts, includeHooks, postData)

		logger.Debugf("new meta after hooks: %v", syndicationMeta)

		for k, v := range syndicationMeta {
			updatePost.FrontMatter[k] = v
		}

		err = dbs.SavePost(updatePost)
		if err != nil {
			logger.Error(err)
			SetFlash(w, "flash", fmt.Sprintf(
				"Your post was saved, but some syndication links might be missing (%v)",
				err))
		}

		err = repo.SavePostFile(updatePost)
		if err != nil {
			logger.Error(err)
			SetFlash(w, "flash", fmt.Sprintf(
				"Your post was saved, but some syndication links might be missing on disk (%v)",
				err))
		}

		// redirect(w, config.TemplatesDir, post.PermaLink())
		http.Redirect(w, r, post.PermaLink(), http.StatusFound)
		return
	}
}

func CreateDeletePostFunc(
	config Config, dbs DBStorage, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating delete post handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		postID := r.PostFormValue("postID")
		logger.Infof("Delete post: %s", postID)

		post, err := dbs.GetPost(postID)
		if err != nil {
			logger.Errorf("Could not find post to delete: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Could not find post to delete: %v", err))
			http.Redirect(w, r, "/edit/"+postID, http.StatusSeeOther)
		}
		logger.Debugf("post: %s date: %s", post.Title, post.PostDate.Format(POSTTIMESTAMPFMT))

		err = dbs.DeletePost(postID)
		if err != nil {
			logger.Errorf("Could not delete post: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Could not delete post: %v", err))
			http.Redirect(w, r, "/edit/"+postID, http.StatusSeeOther)
		}

		err = repo.DeletePostFile(post)
		if err != nil {
			logger.Errorf("Could not delete post file: %v", err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		// redirect(w, config.TemplatesDir, "/")
	}
}
