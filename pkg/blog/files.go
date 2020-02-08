package blog

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	yaml "gopkg.in/yaml.v2"
)

const (
	TAGLISTRE string = `[\s]*,[\s]*`
	HASHTAGRE string = `(?:\s|\A)#[[:alnum:]]+`
)

type PostsRepo interface {
	ListPostFiles() []string
	SavePostFile(post *Post) error
	DeletePostFile(post *Post) error
}

type FilePostsRepo struct {
	PostsDirectory string
}

func (repo *FilePostsRepo) ListPostFiles() []string {
	logger.Debugf("listing files in %s", repo.PostsDirectory)

	files := make([]string, 0)
	files, err := filepath.Glob(filepath.Join(
		repo.PostsDirectory, "*.md"))

	if err != nil {
		logger.Error(err)
	}

	return files
}

func (repo *FilePostsRepo) SavePostFile(post *Post) error {
	filename := fmt.Sprintf(
		"%s-%s.md",
		post.PostDate.Format(POSTDATEFMT),
		post.Slug,
	)

	file := filepath.Join(repo.PostsDirectory, filename)
	logger.Debugf("Write file: %s", file)

	err := ioutil.WriteFile(file, []byte(post.ToString()), 0777)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (repo *FilePostsRepo) DeletePostFile(post *Post) error {
	logger.Debugf("%v", post.PostDate.Format(POSTTIMESTAMPFMT))
	filename := fmt.Sprintf(
		"%s-%s.md",
		post.PostDate.Format(POSTDATEFMT),
		post.Slug,
	)

	file := filepath.Join(repo.PostsDirectory, filename)
	logger.Debugf("Delete file: %s", file)

	err := os.Remove(file)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

/*
GetFrontMatter does a simple key: value parse on the
"yaml" at the front of a post.
*/
func GetFrontMatter(frontmatter string) map[string]string {
	requiredKeys := []string{
		"title", "slug", "date",
	}

	// re := regexp.MustCompile(fmt.Sprintf(`(?i)^(.*?):(.*)$`))

	var fm = make(map[string]string)

	yaml.Unmarshal([]byte(frontmatter), &fm)

	// for _, line := range strings.Split(frontmatter, "\n") {
	// 	m := re.FindStringSubmatch(line)
	// 	if len(m) > 0 && m[1] != "" {
	// 		// normalize keys to lowercase
	// 		fm[strings.ToLower(strings.TrimSpace(m[1]))] = strings.TrimSpace(m[2])
	// 	}
	// }
	for _, key := range requiredKeys {
		if _, ok := fm[key]; !ok {
			fm[key] = ""
		}
	}
	return fm
}

// GetFrontMatterItem scans the yaml post header and
// looks for a key matching `item`
// func GetFrontMatterItem(frontmatter string, item string) string {
// 	re := regexp.MustCompile(fmt.Sprintf(`(?i)^%s:(.*)$`, item))

// 	for _, line := range strings.Split(frontmatter, "\n") {
// 		m := re.FindStringSubmatch(line)
// 		if len(m) > 0 && m[1] != "" {
// 			return strings.TrimSpace(m[1])
// 		}
// 	}
// 	return ""
// }

// read a markdown file with frontmatter into a Post
func ParseFile(path string) (Post, error) {
	content, err := ioutil.ReadFile(path)

	filename := filepath.Base(path)

	var post = NewPost(PostOpts{})

	if err != nil {
		logger.Error(err)
		return post, err
	}

	fileParts := splitFile(string(content))

	if len(fileParts) < 2 {
		logger.Errorf("%q", fileParts)
		return post, errors.New("Bad file format")
	}

	frontMatterStr := fileParts[0]
	frontMatter := GetFrontMatter(frontMatterStr)

	post.FrontMatter = frontMatter
	// logger.Debug(frontMatter)
	body := fileParts[1]
	// logger.Debug(body)

	slug := frontMatter["slug"]
	// if slugIface, ok := frontMatter["slug"]; ok {
	// 	slug := slugIface
	// }
	title := frontMatter["title"]

	dateStr := frontMatter["date"]
	date, err := getPostDate(dateStr, filename)

	if slug == "" {
		slug = getPostSlugFromFile(filename)
	}
	post.Slug = slug
	post.PostDate = date
	post.Title = title

	body = strings.TrimSpace(body)
	post.Body = body

	var tagList []string
	if tagStr, ok := frontMatter["tags"]; ok {
		tagList = splitTags(tagStr)
	} else {
		tagList = []string{}
	}

	var tags []string
	for _, t := range tagList {
		tags = append(tags, strings.TrimSpace(t))
	}

	// add post hashtags, cause that's cool
	processedBody := fmt.Sprintf("%s", markDowner(post.Body))
	hashtags := GetHashTags(processedBody)

	fmt.Printf("Found hashtags: %v", hashtags)
	for _, t := range hashtags {
		if !tagInTags(t, tags) {
			tags = append(tags, t)
		}
	}

	post.Tags = tags

	// logger.Debugf("%q", post)

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
	if dateStr != "" {
		date, err := dateparse.ParseAny(dateStr)

		if err == nil {
			return date, nil
		}
		logger.Warn(err)
		return time.Time{}, err
	}

	logger.Debug("No date found in header...")

	pathRe := regexp.MustCompile(`^([\d]{4})-([\d]{2})-([\d]{2})-`)
	r := pathRe.FindSubmatch([]byte(filename))
	if len(r) == 0 {
		return time.Time{}, errors.New(fmt.Sprintf(
			"Cannot get postdate from dateStr or filename: %s",
			filename))
	}

	year := r[1]
	month := r[2]
	day := r[3]

	dateStr = fmt.Sprintf("%s-%s-%s", year, month, day)
	return time.Parse(POSTDATEFMT, dateStr)
}

func getPostSlugFromFile(filename string) string {
	pathRe := regexp.MustCompile(`^([\d]{4})-([\d]{2})-([\d]{2})-(.*?)\.md`)
	r := pathRe.FindSubmatch([]byte(filename))
	if len(r) != 5 { // r[0] is the full string
		return ""
	}
	slug := string(r[4])
	fmt.Println(slug)
	return slug
}

func tagInTags(tag string, tags []string) bool {
	for _, t := range tags {
		if strings.ToLower(t) == strings.ToLower(tag) {
			return true
		}
	}
	return false
}

func splitTags(tags string) []string {
	re := regexp.MustCompile(TAGLISTRE)
	tagList1 := re.Split(tags, -1)
	var tagList2 []string
	for _, t := range tagList1 {
		if t != "" {
			tagList2 = append(tagList2, strings.ToLower(strings.TrimSpace(t)))
		}
	}
	return tagList2
}

func GetHashTags(s string) []string {
	re := regexp.MustCompile(HASHTAGRE)
	res := re.FindAll([]byte(s), -1)
	var hashtags []string
	for _, b := range res {
		m := string(b)
		tag := strings.Trim(strings.TrimSpace(m), "#")
		hashtags = append(hashtags, tag)
	}
	return hashtags
}

// func linkTags(s string, urlPrefix string) string {
// 	re := regexp.MustCompile(HASHTAGRE)
// 	res := re.([]byte(s), -1)
// 	var hashtags []string
// 	for _, b := range res {
// 		hashtags = append(hashtags, strings.ToLower(
// 			strings.Trim(string(b), "#")))
// 	}
// 	return hashtags
// }

func MakePostSlug(title string) string {
	bits := strings.Split(title, " ")
	s := strings.Join(bits, "-")
	s = strings.ToLower(s)
	re := regexp.MustCompile("[^[:alnum:]-]")
	// leave only characters and dashes
	s = string(re.ReplaceAll([]byte(s), []byte("")))
	s = strings.TrimRight(s, "-")
	return s
}

func MakeNoteSlug(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	str := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("txt-%s", str[0:7])
}
