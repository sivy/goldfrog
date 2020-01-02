package blog

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/leekchan/gtf"
)

// CreateIndexFunc
func CreateIndexFunc(config Config, db *sql.DB, templatesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Serving index...")
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

		isOwner := checkIsOwner(config, r)

		err = t.ExecuteTemplate(w, "base", struct {
			Posts      []Post
			Config     Config
			IsOwner    bool
			FormAction string
			TextHeight int
			ShowSlug   bool
			ShowExpand bool
		}{
			Posts:      posts,
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

func CreateNewPostFunc(
	config Config, db *sql.DB, templatesDir string, repo PostsRepo) http.HandlerFunc {
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
				FormAction string
				IsOwner    bool
				TextHeight int
				ShowSlug   bool
				ShowExpand bool
			}{
				Config:     config,
				FormAction: "/new",
				IsOwner:    true,
				TextHeight: 20,
				ShowSlug:   true,
				ShowExpand: false,
			})
			return
		}

		title := r.PostFormValue("title")
		tags := r.PostFormValue("tags")
		body := r.PostFormValue("body")

		p := Post{
			Title:    title,
			Tags:     strings.Split(tags, ","),
			Body:     body,
			Slug:     makePostSlug(title),
			PostDate: time.Now(),
		}

		log.Debug(p)
		err := repo.SavePostFile(p)
		if err != nil {
			log.Errorf("Could not save post file: %v", err)
		}

		err = CreatePost(db, p)
		if err != nil {
			log.Errorf("Could not save post: %v", err)
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
		return
	}
}

func CreateSigninPageFunc(
	config Config, dbFile string, templatesDir string) http.HandlerFunc {
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
			t, err := getTemplate(templatesDir, "redirect.html")
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
	s := markdown.ToHTML(
		[]byte(fmt.Sprintf("%s", args...)), nil, nil)
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
	}).Funcs(gtf.GtfFuncMap)

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
	t, err := getTemplate(templatesDir, "redirect.html")
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
