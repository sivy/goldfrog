package webmention

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestGetLinkEndpoint(t *testing.T) {
	client := NewWebMentionClient()

	linkHtml := `
	<link href="http://aaronpk.example/webmention-endpoint" rel="webmention" />
	`
	linkDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(linkHtml))
	linkEndpoint := client.getHtmlEndpoint(linkDoc, []string{"link"})

	assert.Equal(t, linkEndpoint, "http://aaronpk.example/webmention-endpoint")
}

func TestGetAnchorEndpoint(t *testing.T) {
	client := NewWebMentionClient()

	aHtml := `
	<a href="http://aaronpk.example/webmention-endpoint" rel="webmention">webmention</a>
	`
	aDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(aHtml))

	aEndpoint := client.getHtmlEndpoint(aDoc, []string{"a"})

	assert.Equal(t, aEndpoint, "http://aaronpk.example/webmention-endpoint")
}

func TestFindLinks(t *testing.T) {
	htmlStr := `
	<a href="http://example.com">Example Site</a>
	<a href="http://google.com">Google Site</a>
	<a href="http://webmentions.rocks">Webmentions Rocks</a>
	`
	client := NewWebMentionClient()

	links, _ := client.FindLinks(htmlStr)

	assert.Equal(t, 3, len(links), "Found 4 links in the htmlStr")
	for _, href := range []string{
		"http://example.com", "http://google.com", "http://webmentions.rocks",
	} {
		assert.Contains(t, links, href, fmt.Sprintf("found expected %s in links", href))
	}
}
