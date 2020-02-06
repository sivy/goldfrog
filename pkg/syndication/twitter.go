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
}

func (tp *TwitterPoster) formatMessage(postData *PostData) string {
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

func (tp *TwitterPoster) HandlePost(postData *PostData) {
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

	var content = tp.formatMessage(postData)

	// if !linkOnly {
	// 	content = post.Body
	// }
	opts := MicroMessageOpts{
		MaxLength: twitterMaxMessageLen,
	}

	if postData.Title != "" {
		opts.Title = postData.Title
		opts.PermaLink = tp.BaseUrl + postData.PermaLink
	} else {
		opts.ShortID = postData.Slug
	}

	content = tp.formatMessage(postData, opts)

	tweet, _, err := client.Statuses.Update(
		content, &twitter.StatusUpdateParams{})

	if err != nil {
		fmt.Printf("%v", err)
	}

	post.FrontMatter["twitter_id"] = tweet.IDStr

	url := fmt.Sprintf(
		"https://twitter.com/%s/status/%s",
		tweet.User.ScreenName,
		tweet.IDStr)

	// should be saved after these return
	post.FrontMatter["twitter_url"] = url

	logger.Debugf("Posted status: %s", url)
}

func (tp *TwitterPoster) LinkForID(config Config, id string) string {
	return fmt.Sprintf(config.Twitter.LinkFormat, config.Twitter.UserID, id)
}

func NewTwitterPoster(opts TwitterOpts) {
	return TwitterPoster{
		ClientKey:    opts.ClientKey,
		ClientSecret: opts.ClientSecret,
		AccessKey:    opts.AccessKey,
		AccessSecret: opts.AccessSecret,
	}
}
