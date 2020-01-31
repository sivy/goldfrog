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
)

const (
	TAGLISTRE string = `[\s]*,[\s]*`
	HASHTAGRE string = `(?:\s|\A)#[[:alnum:]]+`
)

type PostsRepo struct {
	PostsDirectory string
}

func (repo *PostsRepo) ListPostFiles() []string {
	logger.Debugf("listing files in %s", repo.PostsDirectory)

	files := make([]string, 0)
	files, err := filepath.Glob(filepath.Join(
		repo.PostsDirectory, "*.md"))

	if err != nil {
		logger.Error(err)
	}

	return files
}

func (repo *PostsRepo) SavePostFile(post *Post) error {
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

func (repo *PostsRepo) DeletePostFile(post *Post) error {
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
GetFrontMatter soes a simple key: value parse on the
"yaml" at the front of a post.
*/
func GetFrontMatter(frontmatter string) map[string]string {
	requiredKeys := []string{
		"title", "slug", "date",
	}

	re := regexp.MustCompile(fmt.Sprintf(`(?i)^(.*?):(.*)$`))

	var fm = make(map[string]string)

	for _, line := range strings.Split(frontmatter, "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) > 0 && m[1] != "" {
			// normalize keys to lowercase
			fm[strings.ToLower(strings.TrimSpace(m[1]))] = strings.TrimSpace(m[2])
		}
	}
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

	if err != nil {
		logger.Errorf("Could not parse frontmatter! %v", err)
		return NewPost(PostOpts{}), err
	}
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
	hashtags := getHashTags(processedBody)

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
	logger.Debugf("Got dateStr: %s", dateStr)
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

func getHashTags(s string) []string {
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

func makeNoteSlug(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	str := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("txt-%s", str[0:7])
}

type MicroMessageOpts struct {
	MaxLength int
	Title     string
	PermaLink string
	ShortID   string
	NumParas  int
	Tags      []string
}

func makeMicroMessage(
	source string, opts MicroMessageOpts) string {
	/*
		<Title optional

		><Body

		><Tags optional

		><link>
	*/

	if opts.NumParas == 0 {
		opts.NumParas = -1
	}

	source = stripHTML(markDowner(source))
	var title string
	var link string

	if opts.Title != "" {
		title = opts.Title
	}
	if opts.PermaLink != "" {
		link = fmt.Sprintf("(%s)", opts.PermaLink)
	}
	if opts.ShortID != "" {
		link = fmt.Sprintf("(monkinetic %s)", opts.ShortID)
	}

	var fmtTagStr string
	if len(opts.Tags) != 0 {
		var fmtTags []string
		for _, t := range opts.Tags {
			if t == "" {
				continue
			}
			if regexp.MustCompile("#" + t).MatchString(source) {
				continue
			}
			fmtTags = append(fmtTags, fmt.Sprintf("#%s", t))
		}
		fmtTagStr = strings.Join(fmtTags, " ")
	}

	var messageParts []string
	availableChars := opts.MaxLength
	if title != "" {
		availableChars -= len(title) + 2 // len(\n\n)
	}
	if link != "" {
		availableChars -= len(link) + 2 // len(\n\n)
	}
	if fmtTagStr != "" {
		availableChars -= len(fmtTagStr) + 2 // len(\n\n)
	}

	// split paras
	sourceParas := strings.Split(source, "\n\n")
	var messageParas []string
	var messageBody string
	for n, para := range sourceParas {
		if opts.NumParas < 0 {
			if len(strings.Join(messageParas, "\n\n"))+len(para) < availableChars {
				messageParas = append(
					messageParas, strings.TrimSpace(para))
			} else {
				break
			}
		} else if opts.NumParas > 0 && n <= opts.NumParas {
			if len(strings.Join(messageParas, "\n\n"))+len(para) < availableChars {
				messageParas = append(
					messageParas, strings.TrimSpace(para))
			} else {
				break
			}
		}
	}
	messageBody = strings.Join(messageParas, "\n\n")
	// find closes para that fits in available length

	if title != "" {
		messageParts = append(messageParts, title)
	}
	messageParts = append(messageParts, messageBody)

	if link != "" {
		messageParts = append(messageParts, link)
	}
	if fmtTagStr != "" {
		messageParts = append(messageParts, fmtTagStr)
	}

	microMessage := strings.Join(messageParts, "\n\n")
	logger.Debugf("microMessage: %s", microMessage)
	return microMessage
}
