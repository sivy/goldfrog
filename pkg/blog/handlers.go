package blog

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/leekchan/gtf"
)

// CreateIndexFunc
func CreateIndexFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating index handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving index...")

		postOpts := GetPostOpts{}
		getPaginationOpts(r, &postOpts)

		posts := GetPosts(db, postOpts)

		log.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(config.TemplatesDir, "index.html")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		log.Debugf("index template: %v", t)

		isOwner := checkIsOwner(config, r)

		err = t.ExecuteTemplate(w, "base", struct {
			Posts      []Post
			Post       Post
			Config     Config
			IsOwner    bool
			FormAction string
			TextHeight int
			ShowSlug   bool
			ShowExpand bool
		}{
			Posts:      posts,
			Post:       Post{},
			Config:     config,
			IsOwner:    isOwner,
			FormAction: "/new",
			ShowSlug:   false,
			TextHeight: 10,
			ShowExpand: true,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

// CreateIndexFunc
func CreateRssFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating rss handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving index...")

		postOpts := GetPostOpts{Limit: 10}
		posts := GetPosts(db, postOpts)

		log.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(config.TemplatesDir, "base/rss.xml")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8" standalone="yes" ?>`))
		err = t.ExecuteTemplate(w, "rss", struct {
			Posts  []Post
			Config Config
		}{
			Posts:  posts,
			Config: config,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreatePostPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating post detail handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving post page...")

		postSlug := chi.URLParam(r, "slug")

		post, err := GetPostBySlug(db, postSlug)

		log.Debugf("Found post: %s", post.Title)

		if err != nil {
			log.Errorf("Could not get post: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		t, err := getTemplate(config.TemplatesDir, "post_detail.html")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		log.Debug(t)

		isOwner := checkIsOwner(config, r)

		err = t.ExecuteTemplate(w, "base", struct {
			Post    Post
			Config  Config
			IsOwner bool
		}{
			Post:    post,
			Config:  config,
			IsOwner: isOwner,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateArchiveYearMonthFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating main archive list handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving archive year/month...")

		archiveData := GetArchiveYearMonths(db)

		t, err := getTemplate(config.TemplatesDir, "archive_years.html")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		err = t.ExecuteTemplate(w, "base", struct {
			ArchiveData []ArchiveEntry
			Config      Config
		}{
			ArchiveData: archiveData,
			Config:      config,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateArchivePageFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating archive page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving archive year/month...")

		// postOpts := GetPostOpts{Limit: 10}
		year := chi.URLParam(r, "year")
		month := chi.URLParam(r, "month")

		posts := GetArchivePosts(db, year, month)

		t, err := getTemplate(config.TemplatesDir, "archive_posts.html")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		err = t.ExecuteTemplate(w, "base", struct {
			Posts  []Post
			Config Config
			Year   string
			Month  string
		}{
			Posts:  posts,
			Config: config,
			Year:   year,
			Month:  month,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateSearchPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating tag page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving search results...")

		isOwner := checkIsOwner(config, r)

		var posts []Post
		t, err := getTemplate(config.TemplatesDir, "post_list.html")

		// postOpts := GetPostOpts{Limit: 10}
		term := r.PostFormValue("s")
		if term == "" {
			term = r.URL.Query().Get("s")
		}

		if term == "" {
			err = t.ExecuteTemplate(w, "base", struct {
				Posts   []Post
				Config  Config
				Title   string
				IsOwner bool
			}{
				Posts:   posts,
				Config:  config,
				Title:   fmt.Sprintf("No search term!"),
				IsOwner: isOwner,
			})
			return
		}
		opts := GetPostOpts{
			Title: term,
			Body:  term,
		}

		posts = GetPosts(db, opts)
		log.Debugf("found posts: %d", len(posts))

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		err = t.ExecuteTemplate(w, "base", struct {
			Posts   []Post
			Config  Config
			Title   string
			IsOwner bool
		}{
			Posts:   posts,
			Config:  config,
			Title:   fmt.Sprintf("Posts found for '%s'", term),
			IsOwner: isOwner,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateTagPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	log.Debug("Creating tag page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving tag search...")

		// postOpts := GetPostOpts{Limit: 10}
		tag := chi.URLParam(r, "tag")

		posts := GetTaggedPosts(db, tag)
		log.Debugf("tagged posts: %d", len(posts))
		t, err := getTemplate(config.TemplatesDir, "post_list.html")

		if err != nil {
			log.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		isOwner := checkIsOwner(config, r)

		err = t.ExecuteTemplate(w, "base", struct {
			Posts   []Post
			Config  Config
			Title   string
			IsOwner bool
		}{
			Posts:   posts,
			Config:  config,
			Title:   fmt.Sprintf("Posts tagged with '%s'", tag),
			IsOwner: isOwner,
		})

		if err != nil {
			log.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateNewPostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	log.Debug("Creating new post form handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if !checkIsOwner(config, r) {
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			log.Info("Rendering New Post form")
			t, err := getTemplate(config.TemplatesDir, "newpost.html")
			if err != nil {
				log.Errorf("Could not get template: %v", err)
			}

			err = t.ExecuteTemplate(w, "base", struct {
				Config     Config
				Post       Post
				FormAction string
				IsOwner    bool
				TextHeight int
				ShowSlug   bool
				ShowExpand bool
			}{
				Config:     config,
				Post:       Post{},
				FormAction: "/new",
				IsOwner:    true,
				TextHeight: 20,
				ShowSlug:   true,
				ShowExpand: false,
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
			log.Infof("File upload in progress...")

			imagePath = filepath.Join(config.UploadsDir, handler.Filename)
			log.Debugf("Writing uploaded file: %s", imagePath)
			f, err := os.OpenFile(
				imagePath, os.O_WRONLY|os.O_CREATE, 0777)

			if err != nil {
				log.Error(err)
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

		p := Post{
			Title:    title,
			Tags:     splitTags(tags),
			Body:     body,
			Slug:     slug,
			PostDate: time.Now(),
		}

		log.Debug(p)
		p.Tags = updateTags(p.Body, p.Tags)

		err = repo.SavePostFile(p)
		if err != nil {
			log.Errorf("Could not save post file: %v", err)
		}

		err = CreatePost(db, p)
		if err != nil {
			log.Errorf("Could not save post: %v", err)
		}

		crossPosters := MakeCrossPosters(config)

		if r.PostFormValue("twitter") == "on" {
			postedUrl := crossPosters["twitter"].SendPost(p, false)
			if postedUrl != "" {
				log.Infof("Posted twtiter message: %s", postedUrl)
			}
		}

		if r.PostFormValue("mastodon") == "on" {
			postedUrl := crossPosters["mastodon"].SendPost(p, false)
			if postedUrl != "" {
				log.Infof("Posted mastodon message: %s", postedUrl)
			}
		}

		redirect(w, config.TemplatesDir, "/")
		return
	}
}

func CreateEditPostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	log.Debug("Creating edit post handler")
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			log.Info("Rendering Edit Post form")

			postID := chi.URLParam(r, "postID")
			log.Debugf("Post %s", postID)

			post, err := GetPost(db, postID)

			if err != nil {
				log.Error(err)
			}
			log.Debugf("Found post %s", post.Title)

			if !checkIsOwner(config, r) {
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			t, err := getTemplate(config.TemplatesDir, "editpost.html")
			if err != nil {
				log.Errorf("Could not get template: %v", err)
			}
			log.Debug(t)

			err = t.ExecuteTemplate(w, "base", struct {
				Config     Config
				Post       Post
				FormAction string
				IsOwner    bool
				TextHeight int
				ShowSlug   bool
				ShowExpand bool
			}{
				Config:     config,
				Post:       post,
				FormAction: "/edit",
				IsOwner:    true,
				TextHeight: 20,
				ShowSlug:   true,
				ShowExpand: false,
			})
			return
		}
		log.Info("Handling Edit Post form")

		postID := r.FormValue("postID")
		post, err := GetPost(db, postID)
		if err != nil {
			log.Error(err)
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
			log.Infof("File upload in progress...")

			imagePath = filepath.Join(config.UploadsDir, handler.Filename)
			log.Debugf("Writing uploaded file: %s", imagePath)

			f, err := os.OpenFile(
				imagePath, os.O_WRONLY|os.O_CREATE, 0777)
			if err != nil {
				log.Error(err)
				hasImage = false
			} else {
				defer f.Close()
				io.Copy(f, file)
				imageUrl = strings.Join([]string{
					config.Blog.Url, "uploads", handler.Filename}, "/")
			}
		} else {
			log.Error(err)
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

		log.Debug(post)

		err = repo.SavePostFile(post)
		if err != nil {
			log.Errorf("Could not save post file: %v", err)
		}

		err = SavePost(db, post)
		if err != nil {
			log.Errorf("Could not save post: %v", err)
		}

		redirect(w, config.TemplatesDir, post.Url())
		return
	}
}

func CreateDeletePostFunc(
	config Config, db *sql.DB, repo PostsRepo) http.HandlerFunc {
	log.Debug("Creating delete post handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			redirect(w, config.TemplatesDir, "/")
		}

		postID := r.PostFormValue("postID")

		post, err := GetPost(db, postID)

		err = DeletePost(db, postID)
		if err != nil {
			log.Errorf("Could not delete post: %v", err)
		}

		err = repo.DeletePostFile(post)
		if err != nil {
			log.Errorf("Could not delete post file: %v", err)
		}

		redirect(w, config.TemplatesDir, "/")
	}
}

func CreateSigninPageFunc(
	config Config, dbFile string) http.HandlerFunc {
	log.Debug("Creating signin handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			log.Infof("Handle post ")
			user := r.PostFormValue("username")
			pwd := r.PostFormValue("password")

			hash := hashAccount(user, pwd)

			if user != "" && user == config.Signin.Username {
				if pwd != "" && pwd == config.Signin.Password {
					http.SetCookie(w, &http.Cookie{
						Name:    "goldfrog",
						Value:   hash,
						Path:    "/",
						Expires: time.Now().AddDate(1, 0, 0),
						// Secure: true,
					})
				}
			}

			t, err := getTemplate(config.TemplatesDir, "base/redirect.html")
			if err != nil {
				log.Errorf("Could not get template: %v", err)
			}

			err = t.ExecuteTemplate(w, "redirect", struct {
				Config Config
				Url    string
			}{
				Config: config,
				Url:    "/",
			})
			if err != nil {
				log.Warnf("Error rendering... %v", err)
			}
			return
		}

		t, err := getTemplate(config.TemplatesDir, "signin.html")

		if err != nil {
			log.Error(err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		// w.Header().Set("Content-Type", "text/html")
		err = t.ExecuteTemplate(w, "base", struct {
			Config Config
		}{
			Config: config,
		})
		if err != nil {
			log.Warnf("Error rendering... %v", err)
		}
	}
}

func CreateSignoutPageFunc(
	config Config, dbFile string) http.HandlerFunc {
	log.Debug("Creating signout handler")
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "goldfrog",
			Value:   "",
			Path:    "/",
			Expires: time.Now().AddDate(-1, 0, 0),
			// Secure: true,
		})

		redirect(w, config.TemplatesDir, "/")
		return
	}
}

func markDowner(args ...interface{}) template.HTML {
	extensions := parser.CommonExtensions | parser.HeadingIDs
	parser := parser.NewWithExtensions(extensions)
	content := fmt.Sprintf("%s", args...)
	s := markdown.ToHTML(
		[]byte(content), parser, nil)

	return template.HTML(s)
}

func excerpter(args ...interface{}) template.HTML {
	s := fmt.Sprintf("%s", args...)
	s = s[0:255]
	return template.HTML(s)
}

func htmlEscaper(args ...interface{}) string {
	s := fmt.Sprintf("%s", args...)
	return s
}

func hashtagger(args ...interface{}) template.HTML {
	s := fmt.Sprintf("%s", args...)
	re := regexp.MustCompile(`([\s\>])#([[:alnum:]]+)\b`)
	s = re.ReplaceAllString(s, "$1<a href=\"/tag/$2\">#$2</a>")
	return template.HTML(s)
}

func getTemplate(templatesDir string, name string) (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{
		"markdown":  markDowner,
		"excerpt":   excerpter,
		"escape":    htmlEscaper,
		"hashtags":  hashtagger,
		"striphtml": stripHTML,
		// "isOwner": makeIsOwner(isOwner)
	}).Funcs(gtf.GtfFuncMap)

	t, err := t.ParseGlob(filepath.Join(templatesDir, "base/*.html"))
	t, err = t.ParseFiles(
		filepath.Join(templatesDir, name),
	)
	if err != nil {
		return t, err
	}

	if err != nil {
		return t, err
	}

	return t, nil
}

func hashAccount(user string, password string) string {
	log.Debug("hashAccount...")
	h := sha1.New()
	h.Write([]byte(user + password))
	bytes := h.Sum(nil)
	hash := fmt.Sprintf("%x", bytes)
	return hash
}

func checkIsOwner(config Config, r *http.Request) bool {
	checkCookie, err := r.Cookie("goldfrog")
	isOwner := false

	if err == nil {
		checkHash := checkCookie.Value
		accountHash := hashAccount(
			config.Signin.Username, config.Signin.Password)

		if checkHash == accountHash {
			log.Debugf("isOwner: true")
			isOwner = true
		}
	}
	return isOwner
}

func redirect(w http.ResponseWriter, templatesDir string, url string) {
	t, err := getTemplate(templatesDir, "base/redirect.html")
	if err != nil {
		log.Errorf("Could not get template: %v", err)
		return
	}

	err = t.ExecuteTemplate(w, "redirect", struct {
		Url string
	}{
		Url: url,
	})
}

func updateTags(body string, tags []string) []string {
	log.Debugf("Start tags: %q", tags)
	hashtags := getHashTags(body)
	log.Debugf("Found hashtags: %v", hashtags)
	for _, t := range hashtags {
		if !tagInTags(t, tags) {
			tags = append(tags, t)
		}
	}
	log.Debugf("End tags: %v", tags)
	return tags
}

func getPaginationOpts(r *http.Request, opts *GetPostOpts) {
	page := 0
	pageOpt := r.FormValue("page")
	i, err := strconv.Atoi(pageOpt)
	if err != nil {
		log.Debug(err)
	}
	page = i

	offset := 0

	limit := 20
	limitOpt := r.FormValue("limit")
	if limitOpt != "" {
		i, err := strconv.Atoi(limitOpt)
		if err != nil {
			log.Debug(err)
		} else {
			limit = i
		}
	}

	if page != 0 {
		offset = page * limit
	}

	opts.Limit = limit
	opts.Offset = offset
}
