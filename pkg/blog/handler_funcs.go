package blog

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/leekchan/gtf"
)

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

// func makeFlashFunc(w http.ResponseWriter, r *http.Request) func(args ...interface{}) template.HTML {
// 	logger.Debugf("make flash function with writer: %v", w)
// 	return func(args ...interface{}) template.HTML {
// 		flash, _ := GetFlash(w, r, "flash")
// 		if flash != "" {
// 			return template.HTML(
// 				fmt.Sprintf("<div class='flash'>%s</div>", flash),
// 			)
// 		}
// 		return template.HTML("")
// 	}
// }

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
	if err != nil {
		return t, err
	}

	t, err = t.ParseFiles(
		filepath.Join(templatesDir, name),
	)
	if err != nil {
		return t, err
	}

	return t, nil
}

func hashAccount(user string, password string) string {
	logger.Debug("hashAccount...")
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
			logger.Debugf("isOwner: true")
			isOwner = true
		}
	}
	return isOwner
}

// func redirect(w http.ResponseWriter, templatesDir string, url string) {
// 	t, err := getTemplate(templatesDir, "base/redirect.html")
// 	if err != nil {
// 		logger.Errorf("Could not get template: %v", err)
// 		return
// 	}

// 	err = t.ExecuteTemplate(w, "redirect", struct {
// 		Url string
// 	}{
// 		Url: url,
// 	})
// }

func updateTags(body string, tags []string) []string {
	logger.Debugf("Start tags: %q", tags)
	hashtags := getHashTags(body)
	logger.Debugf("Found hashtags: %v", hashtags)
	for _, t := range hashtags {
		if !tagInTags(t, tags) {
			tags = append(tags, t)
		}
	}
	logger.Debugf("End tags: %v", tags)
	return tags
}

func getPaginationOpts(r *http.Request, opts *GetPostOpts) {
	page := 0
	pageOpt := r.FormValue("page")
	if pageOpt != "" {
		i, err := strconv.Atoi(pageOpt)
		if err != nil {
			logger.Debug(err)
		}
		page = i
	}

	offset := 0

	limit := 20
	limitOpt := r.FormValue("limit")
	if limitOpt != "" {
		i, err := strconv.Atoi(limitOpt)
		if err != nil {
			logger.Debug(err)
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