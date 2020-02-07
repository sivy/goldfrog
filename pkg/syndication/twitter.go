package syndication

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const (
	twitterMaxMessageLen int = 280
)

type TwitterPoster struct {
	BaseUrl      string
	ClientKey    string
	ClientSecret string
	AccessKey    string
	AccessSecret string
	LinkFormat   string
	UserID       string
	MaxLen       int
}

func (tp *TwitterPoster) FormatMessage(postData PostData) string {

	source := stripHTML(markDowner(postData.Body))
	var title string
	var link string

	if postData.Title != "" {
		title = postData.Title
	}

	if postData.PermaLink != "" {
		link = fmt.Sprintf("(%s)", postData.PermaLink)
	}
	if postData.ShortID != "" {
		link = fmt.Sprintf("(monkinetic %s)", postData.ShortID)
	}

	var fmtTagStr string
	if len(postData.Tags) != 0 {
		var fmtTags []string
		for _, t := range postData.Tags {
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
	availableChars := tp.MaxLen
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
	var messageBody string

	if len(source) < availableChars {
		// if message fits, do it all
		messageBody = source
	} else {
		// if it doesn't, only do first para
		messageBody = sourceParas[0]
	}

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

func (tp *TwitterPoster) HandlePost(postData PostData) map[string]string {
	logger.Infof("Handling Twitter crosspost...")
	config := oauth1.NewConfig(
		tp.ClientKey,
		tp.ClientSecret)
	token := oauth1.NewToken(
		tp.AccessKey,
		tp.AccessSecret,
	)

	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	var content = tp.FormatMessage(postData)

	tweet, _, err := client.Statuses.Update(
		content, &twitter.StatusUpdateParams{})

	if err != nil {
		fmt.Printf("%v", err)
	}

	var resultData = make(map[string]string)
	resultData["twitter_id"] = tweet.IDStr

	url := fmt.Sprintf(
		"https://twitter.com/%s/status/%s",
		tweet.User.ScreenName,
		tweet.IDStr)

	// should be saved after these return
	resultData["twitter_url"] = url

	logger.Debugf("Post results: %v", resultData)
	return resultData
}

func (tp *TwitterPoster) LinkForID(id string) string {

	return fmt.Sprintf("https://twitter.com/%s/status/%s", tp.UserID, id)
}

func NewTwitterPoster(opts TwitterOpts) *TwitterPoster {
	return &TwitterPoster{
		ClientKey:    opts.ClientKey,
		ClientSecret: opts.ClientSecret,
		AccessKey:    opts.AccessKey,
		AccessSecret: opts.AccessSecret,
		UserID:       opts.UserID,
		MaxLen:       280,
	}
}
