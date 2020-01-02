package blog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type PostsRepo struct {
	PostsDirectory string
}

func (repo *PostsRepo) ListPostFiles() []string {
	log.Debugf("listing files in %s", repo.PostsDirectory)

	files := make([]string, 0)
	files, err := filepath.Glob(filepath.Join(
		repo.PostsDirectory, "*.md"))

	if err != nil {
		log.Error(err)
	}

	return files
}

func (repo *PostsRepo) SavePostFile(post Post) error {
	filename := fmt.Sprintf(
		"%s-%s.md",
		post.PostDate.Format("2006-01-02"),
		post.Slug,
	)

	file := filepath.Join(repo.PostsDirectory, filename)
	log.Debugf("Write file: %s", file)

	err := ioutil.WriteFile(file, []byte(post.ToString()), 0777)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (repo *PostsRepo) DeletePostFile(post Post) error {
	filename := fmt.Sprintf(
		"%s-%s.md",
		post.PostDate.Format("2006-01-02"),
		post.Slug,
	)

	file := filepath.Join(repo.PostsDirectory, filename)
	log.Debugf("Write file: %s", file)

	err := os.Remove(file)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// GetFrontMatterItem scans the yaml post header and
// looks for a key matching `item`
func GetFrontMatterItem(frontmatter string, item string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)^%s:(.*)$`, item))

	for _, line := range strings.Split(frontmatter, "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) > 0 && m[1] != "" {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

// read a markdown file with frontmatter into a Post
func ParseFile(path string) (Post, error) {
	content, err := ioutil.ReadFile(path)

	filename := filepath.Base(path)

	var post Post

	if err != nil {
		log.Error(err)
		return post, err
	}

	fileParts := splitFile(string(content))

	if len(fileParts) < 2 {
		log.Errorf("%q", fileParts)
		return post, errors.New("Bad file format")
	}

	frontMatter := fileParts[0]
	// log.Debug(frontMatter)
	body := fileParts[1]
	// log.Debug(body)

	slug := GetFrontMatterItem(frontMatter, "slug")
	title := GetFrontMatterItem(frontMatter, "title")

	dateStr := GetFrontMatterItem(frontMatter, "date")
	date, err := getPostDate(dateStr, filename)

	if slug == "" {
		slug = getPostSlugFromFile(filename)
	}
	post.Slug = slug
	post.PostDate = date
	post.Title = title

	tagStr := GetFrontMatterItem(frontMatter, "tags")
	tagList := strings.Split(tagStr, ",")
	var tags []string
	for _, t := range tagList {
		tags = append(tags, strings.TrimSpace(t))
	}
	post.Tags = tags

	body = strings.TrimSpace(body)
	post.Body = body

	// log.Debugf("%q", post)

	return post, nil
}

func splitFile(source string) []string {
	sourceBytes := regexp.MustCompile("\r\n").ReplaceAll([]byte(source), []byte("\n"))

	hyphenSep := regexp.MustCompile("\n---\n")
	newlineSep := regexp.MustCompile("\n\n")

	if r := hyphenSep.Find(sourceBytes); r != nil {
		return hyphenSep.Split(string(sourceBytes), -1)
	} else {
		return newlineSep.Split(string(sourceBytes), 2)
	}
}

func getPostDate(dateStr string, filename string) (time.Time, error) {

	date, err := dateparse.ParseAny(dateStr)

	if err == nil {
		return date, nil
	}
	log.Warn(err)

	pathRe := regexp.MustCompile(`^([\d]{4})-([\d]{2})-([\d]{2})-`)
	r := pathRe.FindSubmatch([]byte(filename))
	if len(r) == 0 {
		return time.Now(), errors.New(fmt.Sprintf(
			"Cannot get postdate from dateStr or filename: %s",
			filename))
	}
	year := r[1]
	month := r[2]
	day := r[3]

	dateStr = fmt.Sprintf("%s-%s-%s", year, month, day)
	return time.Parse("2006-01-02", dateStr)
}

func getPostSlugFromFile(filename string) string {
	pathRe := regexp.MustCompile(`^([\d]{4})-([\d]{2})-([\d]{2})-(.*?)\.md`)
	r := pathRe.FindSubmatch([]byte(filename))
	if len(r) != 5 { // r[0] is the full string
		return ""
	}
	fmt.Println(string(r[4]))
	return string(r[4])
}

func makePostSlug(title string) string {
	bits := strings.Split(title, " ")
	s := strings.Join(bits, "-")
	s = strings.ToLower(s)
	re := regexp.MustCompile("[^[:alnum:]-]")
	// leave only characters and dashes
	s = string(re.ReplaceAll([]byte(s), []byte("")))
	s = strings.TrimRight(s, "-")
	return s
}
