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
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/leekchan/gtf"
)

// CreateIndexFunc
func CreateIndexFunc(config Config, db *sql.DB, templatesDir string) http.HandlerFunc {
	log.Debug("Creating index handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving index...")
		log.Debug(templatesDir)

		postOpts := GetPostOpts{Limit: 10}
		posts := GetPosts(db, postOpts)

		log.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(templatesDir, "index.html")

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

func CreatePostPageFunc(config Config, db *sql.DB, templatesDir string) http.HandlerFunc {
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

		t, err := getTemplate(templatesDir, "post_detail.html")

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

func CreateArchiveYearMonthFunc(config Config, db *sql.DB, templatesDir string) http.HandlerFunc {
	log.Debug("Creating main archive list handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving archive year/month...")

		archiveData := GetArchiveYearMonths(db)

		log.Debugf("Found archiveDAta %v", archiveData)

		t, err := getTemplate(templatesDir, "archive_years.html")

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

func CreateArchivePageFunc(config Config, db *sql.DB, templatesDir string) http.HandlerFunc {
	log.Debug("Creating archive page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving archive year/month...")

		// postOpts := GetPostOpts{Limit: 10}
		year := chi.URLParam(r, "year")
		month := chi.URLParam(r, "month")

		posts := GetArchivePosts(db, year, month)

		t, err := getTemplate(templatesDir, "archive_posts.html")

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

func CreateNewPostFunc(
	config Config, db *sql.DB, templatesDir string, repo PostsRepo, staticDir string) http.HandlerFunc {
	log.Debug("Creating new post form handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if !checkIsOwner(config, r) {
				http.Redirect(w, r, "/", http.StatusUnauthorized)
			}

			log.Info("Rendering New Post form")
			t, err := getTemplate(templatesDir, "newpost.html")
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

		if err == nil {
			// there's an image
			hasImage = true
			defer file.Close()
			log.Infof("File upload in progress...")
			f, err := os.OpenFile(
				filepath.Join(staticDir, "images", handler.Filename),
				os.O_WRONLY|os.O_CREATE, 0777)
			if err != nil {
				log.Error(err)
				hasImage = false
			} else {
				defer f.Close()
				io.Copy(f, file)
				imageUrl = filepath.Join("/static/images", handler.Filename)
			}
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")

		if hasImage {
			body = strings.Replace(
				body, "[image]",
				fmt.Sprintf("![%s](%s)", imageUrl, imageUrl),
				-1)
		}

		p := Post{
			Title:    title,
			Tags:     strings.Split(tags, ","),
			Body:     body,
			Slug:     makePostSlug(title),
			PostDate: time.Now(),
		}

		log.Debug(p)
		err = repo.SavePostFile(p)
		if err != nil {
			log.Errorf("Could not save post file: %v", err)
		}

		err = CreatePost(db, p)
		if err != nil {
			log.Errorf("Could not save post: %v", err)
		}

		// t, err := getTemplate(templatesDir, "base/redirect.html")
		// if err != nil {
		// 	log.Errorf("Could not get template: %v", err)
		// }

		// err = t.ExecuteTemplate(w, "redirect", struct {
		// 	Config Config
		// 	Url    string
		// }{
		// 	Config: config,
		// 	Url:    "/",
		// })
		redirect(w, templatesDir, "/")
		return
	}
}

func CreateEditPostFunc(
	config Config, db *sql.DB, templatesDir string, repo PostsRepo, staticDir string) http.HandlerFunc {
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

			t, err := getTemplate(templatesDir, "editpost.html")
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

		if err == nil {
			// there's an image
			hasImage = true
			defer file.Close()
			log.Infof("File upload in progress...")
			f, err := os.OpenFile(
				filepath.Join(staticDir, "images", handler.Filename),
				os.O_WRONLY|os.O_CREATE, 0777)
			if err != nil {
				log.Error(err)
				hasImage = false
			} else {
				defer f.Close()
				io.Copy(f, file)
				imageUrl = filepath.Join("/static/images", handler.Filename)
			}
		} else {
			log.Error(err)
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")

		if hasImage {
			body = strings.Replace(
				body, "[image]",
				fmt.Sprintf("![%s](%s)", imageUrl, imageUrl),
				-1)
		}

		post.Title = title
		post.Tags = strings.Split(tags, ",")
		post.Body = strings.TrimSpace(body)

		log.Debug(post)

		err = repo.SavePostFile(post)
		if err != nil {
			log.Errorf("Could not save post file: %v", err)
		}

		err = SavePost(db, post)
		if err != nil {
			log.Errorf("Could not save post: %v", err)
		}

		redirect(w, templatesDir, "/")
		return
	}
}

func CreateDeletePostFunc(
	config Config, db *sql.DB, templatesDir string, repo PostsRepo) http.HandlerFunc {
	log.Debug("Creating delete post handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			redirect(w, templatesDir, "/")
		}

		postID := r.PostFormValue("postID")

		err := DeletePost(db, postID)

		if err != nil {
			log.Errorf("Could not delete post: %v", err)
		}

		redirect(w, templatesDir, "/")
	}
}

func CreateSigninPageFunc(
	config Config, dbFile string, templatesDir string) http.HandlerFunc {
	log.Debug("Creating signin handler")
	return func(w http.ResponseWriter, r *http.Request) {
		// templatePaths := []string{
		// 	filepath.Join(templatesDir, "default.html"),
		// 	filepath.Join(templatesDir, "signin/*.html"),
		// }
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
			t, err := getTemplate(templatesDir, "base/redirect.html")
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

		t, err := getTemplate(templatesDir, "signin.html")

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

func markDowner(args ...interface{}) template.HTML {
	extensions := parser.CommonExtensions | parser.HardLineBreak
	parser := parser.NewWithExtensions(extensions)

	s := markdown.ToHTML(
		[]byte(fmt.Sprintf("%s", args...)), parser, nil)
	return template.HTML(s)
}

func excerpter(args ...interface{}) template.HTML {
	s := fmt.Sprintf("%s", args...)
	s = s[0:255]
	return template.HTML(s)
}

func getTemplate(templatesDir string, name string) (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{
		"markdown": markDowner,
		"excerpt":  excerpter,
		// "isOwner": makeIsOwner(isOwner)
	}).Funcs(gtf.GtfFuncMap)

	log.Debug(t.DefinedTemplates())

	t, err := t.ParseFiles(
		filepath.Join(templatesDir, name),
	)
	if err != nil {
		return t, err
	}

	t, err = t.ParseGlob(filepath.Join(templatesDir, "base/*.html"))

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
