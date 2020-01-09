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

const (
	TAGLISTRE string = `[\s]*,[\s]*`
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

	body = strings.TrimSpace(body)
	post.Body = body

	tagStr := GetFrontMatterItem(frontMatter, "tags")
	tagList := splitTags(tagStr)
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
	if dateStr != "" {
		date, err := dateparse.ParseAny(dateStr)

		if err == nil {
			return date, nil
		}
		log.Warn(err)
		return time.Time{}, err
	}

	log.Debug("No date found in header...")

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
	log.Debugf("Got dateStr: %s", dateStr)
	return time.Parse("2006-01-02", dateStr)
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
		tagList2 = append(tagList2, strings.ToLower(strings.TrimSpace(t)))
	}
	return tagList2
}

func getHashTags(s string) []string {
	re := regexp.MustCompile("#[[:alnum:]]+")
	res := re.FindAll([]byte(s), -1)
	var hashtags []string
	for _, b := range res {
		hashtags = append(hashtags, strings.ToLower(
			strings.Trim(string(b), "#")))
	}
	return hashtags
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

func makeMicroMessage(
	source string, length int, withTitle string,
	withLink string, withTags []string) string {
	/*
		<Title optional

		><Body

		><Tags optional

		><link>
	*/
	source = stripHTML(markDowner(source))

	if withTitle != "" {
		withTitle += "\n\n"
	}

	if withLink != "" {
		withLink = "\n\n" + withLink

	}

	var fmtTagStr string
	if len(withTags) != 0 {
		var fmtTags []string
		for _, t := range withTags {
			if t == "" {
				continue
			}
			if regexp.MustCompile("#" + t).MatchString(source) {
				continue
			}
			fmtTags = append(fmtTags, fmt.Sprintf("#%s", t))
		}
		fmtTagStr = strings.Join(fmtTags, " ")
		fmtTagStr = "\n\n" + fmtTagStr
	}

	// establish available chars for body
	// length - len(title)
	//        - len(blog.url + post.url)
	availableChars := length - len(withTitle) - len(withLink) - len(fmtTagStr)

	// split paras
	sourceParas := strings.Split(source, "\n\n")
	var messageParas []string
	var messageBody string
	if len(source) < availableChars {
		messageBody = source
	} else {
		for _, para := range sourceParas {
			if len(strings.Join(messageParas, "\n\n"))+len(para) < availableChars {
				messageParas = append(
					messageParas, strings.TrimSpace(para))
			} else {
				break
			}
		}
		messageBody = strings.Join(messageParas, "\n\n")
	}
	// find closes para that fits in available length

	return withTitle + messageBody + fmtTagStr + withLink
}
