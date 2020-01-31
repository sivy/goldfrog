package blog

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
)

func CreateNewPostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating new post form handler")
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

			err = t.ExecuteTemplate(w, "pagebase", struct {
				Config     Config
				Post       Post
				FormAction string
				IsOwner    bool
				TextHeight int
				ShowSlug   bool
				Flash      string
			}{
				Config: config,
				Post: NewPost(PostOpts{
					Title: title,
					Body:  body,
					Slug:  slug,
					Tags:  tags,
				}),
				FormAction: "/new",
				IsOwner:    true,
				TextHeight: 20,
				ShowSlug:   true,
				Flash:      flash,
			})
			return
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
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")
		slug := r.PostFormValue("slug")

		if title == "" {
			slug = makeNoteSlug(body)
		}

		if slug == "" {
			slug = makePostSlug(title)
		}

		body = strings.Replace(body, "\r\n", "\n", -1)

		if hasImage {
			body = strings.Replace(
				body, "[image]",
				fmt.Sprintf("![%s](%s)", imageUrl, imageUrl),
				-1)
		}

		p := NewPost(PostOpts{
			Title:    title,
			Tags:     splitTags(tags),
			Body:     body,
			Slug:     slug,
			PostDate: time.Now(),
		})

		logger.Debug(p)
		p.Tags = updateTags(p.Body, p.Tags)

		err = repo.SavePostFile(&p)
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

		err = CreatePost(db, p)
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

		var wg sync.WaitGroup
		for _, hook := range hooks {
			logger.Debugf("Adding worker for hook %v", hook)
			wg.Add(1)
			go worker(hook, &p, false, &wg)
		}
		wg.Wait()
		http.Redirect(w, r, "/", http.StatusFound)
		// redirect(w, config.TemplatesDir, "/")
		return
	}
}

func CreateEditPostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating edit post handler")
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

			post, err := GetPost(db, postID)

			if err != nil {
				logger.Error(err)
				flash := fmt.Sprintf("An error occurred: %v", err)

				err = t.ExecuteTemplate(w, "base", struct {
					Config     Config
					Post       *Post
					FormAction string
					IsOwner    bool
					TextHeight int
					ShowSlug   bool
					ShowExpand bool
					Flash      string
				}{
					Config:     config,
					Post:       post,
					FormAction: "/edit",
					IsOwner:    true,
					TextHeight: 20,
					ShowSlug:   true,
					ShowExpand: false,
					Flash:      flash,
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
				Config     Config
				Post       *Post
				FormAction string
				IsOwner    bool
				TextHeight int
				ShowSlug   bool
				ShowExpand bool
				Flash      string
			}{
				Config:     config,
				Post:       post,
				FormAction: "/edit",
				IsOwner:    true,
				TextHeight: 20,
				ShowSlug:   true,
				ShowExpand: false,
				Flash:      flash,
			})
			return
		}
		logger.Info("Handling Edit Post form")

		postID := r.FormValue("postID")
		post, err := GetPost(db, postID)
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

		processedBody := fmt.Sprintf("%s", markDowner(post.Body))
		post.Tags = updateTags(processedBody, post.Tags)

		logger.Debug(post)

		err = repo.SavePostFile(post)
		if err != nil {
			logger.Errorf("Could not save post file: %v", err)
		}

		err = SavePost(db, post)
		if err != nil {
			logger.Errorf("Could not save post: %v", err)
		}

		post.Body = "Updated: " + post.Body

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

		var wg sync.WaitGroup
		for _, hook := range hooks {
			logger.Debugf("Adding worker for hook %v", hook)
			wg.Add(1)
			go worker(hook, post, false, &wg)
		}
		wg.Wait()

		// redirect(w, config.TemplatesDir, post.PermaLink())
		http.Redirect(w, r, post.PermaLink(), http.StatusFound)
		return
	}
}

func CreateDeletePostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	logger.Debug("Creating delete post handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		postID := r.PostFormValue("postID")
		logger.Infof("Delete post: %s", postID)

		post, err := GetPost(db, postID)
		if err != nil {
			logger.Errorf("Could not find post to delete: %v", err)
			SetFlash(w, "flash", fmt.Sprintf("Could not find post to delete: %v", err))
			http.Redirect(w, r, "/edit/"+postID, http.StatusSeeOther)
		}
		logger.Debugf("post: %s date: %s", post.Title, post.PostDate.Format(POSTTIMESTAMPFMT))

		err = DeletePost(db, postID)
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

func worker(hook CrossPoster, post *Post, linkOnly bool, wg *sync.WaitGroup) {
	defer wg.Done()
	postedUrl := hook.SendPost(post, linkOnly)
	if postedUrl != "" {
		logger.Infof("Posted message: %s", postedUrl)
	}
}
