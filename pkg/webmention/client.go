package webmention

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/tomnomnom/linkheader"
)

var logger = logrus.New()

/*
	Interface for WebmentionClient, defines basic functionality
	that differet implementations must provide.
*/
type WebmentionClient interface {
	Fetch(url string) (*http.Response, error)
	EndpointDiscovery(mentionTarget string) (string, error)
	SendWebMentions(source string, links []string)
	SendMention(endpoint string, source string, target string)
	FindLinks(htmlStr string) ([]string, error)
	GetHtmlEndpoint(doc *goquery.Document, elements []string) string
	GetMention(targetUrl string, r io.Reader) (WebMention, error)
}

/*
Client implements the WebMentionClient interface and Webmention client
protocol and allows the system to discover the WebMention endpoint for
a link and send webmentions to the that endpoint. It also provides
functions to parse the links from a snippet of html, and to send
webmentions for a list of links.
*/
type Client struct {
}

func (c *Client) Fetch(url string) (*http.Response, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	logger.Infof("Fetching URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
		return nil, errors.New(
			fmt.Sprintf("No actionable response (%d)", resp.StatusCode))
	}
	return resp, nil
}

// 3.1.2 Sender discovers receiver Webmention endpoint
func (c *Client) EndpointDiscovery(mentionTarget string) (string, error) {

	baseUrl, err := url.Parse(mentionTarget)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	resp, err := c.Fetch(mentionTarget)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	logger.Infof("resp: %s, err: %s", resp.Status, err)

	var endpointValue string
	/*
		Check the HTTP Headers for the first Link header:

		Link: <http://aaronpk.example/webmention>; rel="webmention"
	*/
	endpointValue = c.getHeadEndpoint(resp.Header)

	if endpointValue == "" {
		// test #15, link with empty href resolves to the page url
		// does this count for Link: headers?
		return mentionTarget, nil
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	/*
		Check the HTML for the first link or anchor with rel=webmention

		<link href="http://aaronpk.example/webmention" rel="webmention">
	*/
	if endpointValue == "<none>" {
		endpointValue = c.GetHtmlEndpoint(doc, []string{"link", "a"})
	}

	// nothing found in GetHtmlEndpoint returns "<none>"
	if endpointValue == "" {
		// test #15 resolve empty href to the page url
		return mentionTarget, nil
	}

	if endpointValue == "<none>" {
		return endpointValue, errors.New("No endpoint found")
	}

	endpointUrl, err := url.Parse(endpointValue)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	endpointUrl = baseUrl.ResolveReference(endpointUrl)

	return endpointUrl.String(), nil
}

func (c *Client) getHeadEndpoint(header http.Header) string {
	var endpointValue = "<none>"
	if linkVals, ok := header["Link"]; ok {
		// Get all <link> elements from the <head>
		links := linkheader.ParseMultiple(linkVals)
		for _, link := range links {
			// is there a webmention link?
			if strings.Contains(link.Rel, "webmention") {
				rels := strings.Split(link.Rel, " ")
				for _, rel := range rels {
					if rel == "webmention" {
						return link.URL
					}
				}
			}
		}
	}
	return endpointValue
}

func (c *Client) GetHtmlEndpoint(doc *goquery.Document, elements []string) string {
	var hrefValue = "<none>"

	var selectors []string
	for _, e := range elements {
		selectors = append(selectors, fmt.Sprintf("%s[rel~=webmention]", e))
	}
	var selector = strings.Join(selectors, ",")

	doc.Find(selector).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			value, ok := s.Attr("href")
			if ok {
				hrefValue = value
				return false
			}
			return true
		})

	return hrefValue
}

func (c *Client) FindLinks(htmlStr string) ([]string, error) {
	htmlDoc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	var links []string

	htmlDoc.Find("a[href]").Each(
		func(i int, s *goquery.Selection) {
			link, ok := s.Attr("href")
			if ok {
				links = append(links, link)
			}
		})
	return links, nil
}

/*
SendMention implements the Webmention requirement:
3.1.3 Sender notifies receiver

- Set the form data for the source (post which links to the target)
  and target (linked page)
- Send a POST to the target's webmention endpoint with the data
*/
func (c *Client) SendMention(endpoint string, source string, target string) {
	logger.Infof(
		"Sending webmention (from %s -> %s) to %s",
		source, target, endpoint)

	form := url.Values{}
	form.Add("source", source)
	form.Add("target", target)

	resp, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))

	if err != nil {
		logger.Error(err)
	}

	switch resp.StatusCode {
	case 201:
		{
			logger.Infof(
				"Webmention created, status available at: %s",
				resp.Header.Get("Location"))
		}
	case 202:
		{
			logger.Infof("Webmention accepted")
		}
	default:
		{
			b, _ := ioutil.ReadAll(resp.Body)
			body := string(b)
			logger.Infof("Endpoint returned status: %d, %s", resp.StatusCode, body)
		}
	}
}

// SendWebMentions
func (c *Client) SendWebMentions(source string, links []string) {
	// logger.Infof("Sending webmentions for links %v", links)

	for _, link := range links {
		logger.Debugf("Getting endpoint for link %s", link)
		endpoint, err := c.EndpointDiscovery(link)
		logger.Debugf("Found endpoint %s for link %s", endpoint, link)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debugf("Sending mention for link %v", link)
		c.SendMention(endpoint, source, link)
	}
}

func NewWebMentionClient() WebmentionClient {
	return &Client{}
}
