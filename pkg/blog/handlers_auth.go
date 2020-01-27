package blog

import (
	"net/http"
	"time"
)

func CreateSigninPageFunc(
	config Config, dbFile string) http.HandlerFunc {
	logger.Debug("Creating signin handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			logger.Infof("Handle post ")
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

			t, err := getTemplate(config.TemplatesDir, "base/redirect.html", w, r)
			if err != nil {
				logger.Errorf("Could not get template: %v", err)
			}

			err = t.ExecuteTemplate(w, "redirect", struct {
				Config Config
				Url    string
			}{
				Config: config,
				Url:    "/",
			})
			if err != nil {
				logger.Warnf("Error rendering... %v", err)
			}
			return
		}

		t, err := getTemplate(config.TemplatesDir, "signin.html", w, r)

		if err != nil {
			logger.Error(err)
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
			logger.Warnf("Error rendering... %v", err)
		}
	}
}

func CreateSignoutPageFunc(
	config Config, dbFile string) http.HandlerFunc {
	logger.Debug("Creating signout handler")
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "goldfrog",
			Value:   "",
			Path:    "/",
			Expires: time.Now().AddDate(-1, 0, 0),
			// Secure: true,
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// redirect(w, config.TemplatesDir, "/")
		return
	}
}
