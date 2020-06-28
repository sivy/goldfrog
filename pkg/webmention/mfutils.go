package webmention

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andyleap/microformats"
	"github.com/araddon/dateparse"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

// Get the WebMention (author, content, datetime, etc)
// from the source document
func (c *Client) GetMention(targetUrl string, r io.Reader) (WebMention, error) {
	/*
		TODO: find h-entry
		(microformats parser)
	*/
	var wm = WebMention{
		Type:      "mention",
		Published: time.Now(),
	}

	hentry := getSourceHEntry(targetUrl, r)
	logger.Infof("Found h-entry: %v", hentry)
	if hentry == nil {
		err := errors.New(fmt.Sprintf("Could not find h-entry at target: %s", targetUrl))
		logger.Error(err)
		return WebMention{}, err
	}
	wm.SourceHTML = hentry.HTML

	/*
		TODO: get mention postdate
	*/
	postDate := getHEntryPublished(hentry)
	wm.Published = postDate

	/*
		TODO: get hyperlink to the original post
		(GoQuery)
	*/
	// sourceLink, err := getHEntrySourceLink(hentry.HTML, targetUrl)

	/*
		TODO: if rel="in-reply-to" it's a comment
		(GoQuery?)
	*/
	wm.Type = getHEntryType(hentry, targetUrl)

	/*
		TODO: get author h-card for post,
		or page
	*/
	var author WebMentionAuthor
	if len(hentry.Properties["author"]) > 0 {
		author = WebMentionAuthor{}
		authorMF := hentry.Properties["author"][0].(*microformats.MicroFormat)

		if len(authorMF.Properties["name"]) > 0 {
			author.Name = authorMF.Properties["name"][0].(string)
		}
		if len(authorMF.Properties["url"]) > 0 {
			author.Url = authorMF.Properties["url"][0].(string)
		}
		if len(authorMF.Properties["photo"]) > 0 {
			author.Photo = authorMF.Properties["photo"][0].(string)
		}
		if len(authorMF.Properties["note"]) > 0 {
			author.Note = authorMF.Properties["note"][0].(string)
		}
	}
	wm.Author = author

	/*
		TODO: Get mention content
	*/
	var contentHTML string
	var contentText string

	if len(hentry.Properties["content"]) > 0 {
		contentData := hentry.Properties["content"][0]
		contentHTML = contentData.(map[string]string)["html"]
		contentText = contentData.(map[string]string)["value"]
	}
	wm.ContentText = contentText

	/*
		TODO: sanitize content
	*/
	p := bluemonday.UGCPolicy()
	sanitizedContent := p.Sanitize(contentHTML)
	wm.ContentHTML = sanitizedContent

	/*
		TODO: mention data:
		- source h-entry
		- author h-card
		- post date
		- mention content, sanitized
	*/
	return wm, nil
}

// Get the first h-entry in the source document
// - it *should* be the linking entry if done right
func getSourceHEntry(targetUrl string, r io.Reader) *microformats.MicroFormat {
	parser := microformats.New()
	urlparsed, _ := url.Parse(targetUrl)
	data := parser.Parse(r, urlparsed)

	for _, item := range data.Items {
		if item.Type[0] == "h-entry" {
			return item
		}
	}
	return nil
}

// getHEntryPublished gets the published time
func getHEntryPublished(hentry *microformats.MicroFormat) time.Time {
	postDate := time.Now()
	if len(hentry.Properties["published"]) > 0 {
		parsedDate, err := dateparse.ParseAny(hentry.Properties["published"][0].(string))
		if err == nil {
			postDate = parsedDate
		}
	}
	return postDate
}

// Get hyperlink to the original post
func getHEntrySourceLink(htmlStr string, targetUrl string) (*goquery.Selection, error) {
	node, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		logger.Errorf("Could not parse hentry HTML: %v", htmlStr)
		return nil, err
	}

	doc := goquery.NewDocumentFromNode(node)

	var sourceLink *goquery.Selection
	doc.Find("a[href]").EachWithBreak(
		// func() shouldContinue
		func(i int, s *goquery.Selection) bool {
			link, ok := s.Attr("href")
			if ok {
				if link == targetUrl {
					sourceLink = s
					return false
				}
			}
			return true
		})

	return sourceLink, nil
}

// getHEntryType figures out what kind of comment this is.
// See: https://indieweb.org/post-type-discovery
func getHEntryType(hentry *microformats.MicroFormat, targetUrl string) string {
	println("getHEntryType")

	if len(hentry.Properties) == 0 {
		return "note"
	}

	/*
		TODO: if "in-reply-to" property it's a comment
		(GoQuery?)
	*/
	if isPostType(hentry, "in-reply-to", MakeEqualValidator(targetUrl)) {
		return "comment"
	}

	/*
			If the post has an "rsvp" property with a valid value,
		    Then it is an RSVP post.
	*/
	if isPostType(hentry, "rsvp", RsvpValidator) {
		return "rsvp"
	}

	/*
		If the post has an "in-reply-to" property with a valid URL,
		Then it is a reply post.
	*/
	if isPostType(hentry, "in-reply-to", UrlValidator) {
		return "reply"
	}

	/*
		If the post has a "repost-of" property with a valid URL,
		Then it is a repost (AKA "share") post.
	*/
	if isPostType(hentry, "repost-of", UrlValidator) {
		return "share"
	}

	/*
		If the post has a "like-of" property with a valid URL,
		Then it is a like (AKA "favorite") post.
	*/
	if isPostType(hentry, "like-of", UrlValidator) {
		return "favorite"
	}

	/*
		If the post has a "video" property with a valid URL,
		Then it is a video post.
	*/
	if isPostType(hentry, "video", UrlValidator) {
		return "video"
	}

	/*
		If the post has a "photo" property with a valid URL,
		Then it is a photo post.
	*/
	if isPostType(hentry, "photo", UrlValidator) {
		return "photo"
	}

	/*
		If the post has a "content" property with a non-empty value,
		Then use its first non-empty value as the content
		Else if the post has a "summary" property with a non-empty value,
		Then use its first non-empty value as the content
		Else it is a note post.
		If the post has no "name" property
		  or has a "name" property with an empty string value (or no value)
		Then it is a note post.
		Take the first non-empty value of the "name" property
		Trim all leading/trailing whitespace
		Collapse all sequences of internal whitespace to a single space (0x20) character each
		Do the same with the content
		If this processed "name" property value is NOT a prefix of the processed content,
		Then it is an article post.
		It is a note post.
	*/
	return "note"
}

// isPostType searches the hentry to see if there is a url with
// the given type
func isPostType(
	hentry *microformats.MicroFormat,
	postType string,
	validator func(value string) bool) bool {
	// no properties no dice

	if hentry.Properties[postType] == nil {
		return false
	}
	// Properties is a map[string][]interface{}
	// walk the slice of values of the postType property
	// which are interface{}
	for _, value := range hentry.Properties[postType] {
		logger.Debugf("property value: %v", value)
		if v, ok := value.(string); ok {
			if validator(v) {
				return true
			}
		}
		if v, ok := value.(*microformats.MicroFormat); ok {
			// dataMF := v["data"].(*microformats.MicroFormat)
			if validator(v.Value) {
				return true
			}
			if len(v.Properties["url"]) > 0 &&
				validator(v.Properties["url"][0].(string)) {
				return true
			}
		}
	}

	return false
}

func RsvpValidator(value string) bool {
	for _, v := range []string{
		"yes", "no", "maybe", "interested",
	} {
		if value == v {
			return true
		}
	}
	return false
}

func UrlValidator(value string) bool {
	value = strings.TrimSpace(value)
	urlparsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return urlparsed.Scheme == "https" || urlparsed.Scheme == "http"
}

func MakeEqualValidator(matchValue string) func(value string) bool {
	return func(value string) bool {
		return value == matchValue
	}
}
