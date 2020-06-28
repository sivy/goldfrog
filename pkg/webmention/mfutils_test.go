package webmention

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlValidator(t *testing.T) {
	assert.True(t, UrlValidator("http://example.com/post"), "http://example.com/post is not valid!")
	assert.False(t, UrlValidator("foo://example.com/post"), "foo://example.com/post is not valid!")
	// assert.False(t, UrlValidator("http:///post"), "http:///post is valid!")
	// assert.False(t, UrlValidator("example.com/post"), "example.com/post is valid!")
	// assert.False(t, UrlValidator("foo"), "foo is valid!")
}

func TestHEntrySourceLink(t *testing.T) {
	html := `
	<p class="h-entry">
		<a href="http://target.example">source link</a>
	</p>
	`
	expected := "http://target.example"
	sourceLink, err := getHEntrySourceLink(html, expected)
	assert.Nil(t, err)

	actual, ok := sourceLink.Attr("href")

	assert.True(t, ok)
	assert.Equal(t, actual, expected)
}

func TestHEntry(t *testing.T) {
	htmlStr := `
	<div class="h-entry ">
		<p class="u-in-reply-to">
			<a href="http://target.example">source link</a>
		<p>
	</div>
	`
	expected := "http://target.example"
	hentry := getSourceHEntry(expected, strings.NewReader(htmlStr))
	j, _ := json.MarshalIndent(hentry, "", "  ")
	println(string(j))
	// assert.Equal(t, wmType, "comment")
}

type CheckTestOpts struct {
	shouldSucceed bool
	postType      string
	HTML          string
	validator     func(string) bool
}

func TestHEntryType(t *testing.T) {
	var params = []CheckTestOpts{
		// bad source and target the same
		{
			true,
			"in-reply-to",
			`
				<div class="h-entry">
					<div class="u-in-reply-to h-cite">
						<a class="p-name u-url" href="http://example.com/post">Example Post</a>
					</div>
				</div>
			`,
			MakeEqualValidator("http://example.com/post"),
		},
		{
			true,
			"in-reply-to",
			`
				<div class="h-entry">
					<div class="u-in-reply-to h-cite">
						<a class="p-name u-url" href="http://example.com/post2">Example Post</a>
					</div>
				</div>
			`,
			UrlValidator,
		},
		{
			true,
			"rsvp",
			`
				<div class="h-entry">
					<div>
						<span class="p-rsvp">yes</span> to <a class="p-name u-url" href="http://example.com/post">Example Post</a>
					</div>
				</div>
			`,
			RsvpValidator,
		},
		{
			true,
			"rsvp",
			`
			<div class="h-entry">
				<div>
				<span class="p-rsvp">maybe</span> to <a class="p-name u-url" href="http://example.com/post">Example Post</a>
				</div>
			</div>
			`,
			RsvpValidator,
		},
		// repost-of
		{
			true,
			"in-reply-to",
			`
				<div class="h-entry">
					<div class="u-in-reply-to h-cite">
						<a class="p-name u-url" href="http://example.com/post2">Example Post</a>
					</div>
				</div>
			`,
			UrlValidator,
		},
		{
			true,
			"repost-of",
			`
				<div class="h-entry">
					<div class="u-repost-of h-cite">
						<a class="p-name u-url" href="http://example.com/post2">Example Post</a>
					</div>
				</div>
			`,
			UrlValidator,
		},
		{
			true,
			"like-of",
			`
				<div class="h-entry">
					<div class="u-like-of h-cite">
						<a class="p-name u-url" href="http://example.com/post2">Example Post</a>
					</div>
				</div>
			`,
			UrlValidator,
		},
		{
			true,
			"video",
			`
				<div class="h-entry">
					<div class="u-video h-cite">
						<a class="p-name u-url" href="http://example.com/video">Example Post</a>
					</div>
				</div>
			`,
			UrlValidator,
		},
		{
			true,
			"photo",
			`
				<div class="h-entry">
					<img class="u-photo u-url" src="http://example.com/photo.jpeg" />
				</div>
			`,
			UrlValidator,
		},
	}

	expected := "http://example.com/post"

	for _, opts := range params {
		fmt.Printf("type: %s shouldSucceed: %v\n", opts.postType, opts.shouldSucceed)
		hentry := getSourceHEntry(expected, strings.NewReader(opts.HTML))
		hentryJson, _ := json.MarshalIndent(hentry, "", "  ")
		success := isPostType(hentry, opts.postType, opts.validator)
		assert.Equal(
			t,
			opts.shouldSucceed,
			success,
			string(hentryJson),
		)
		typeResult := getHEntryType(hentry, expected)
		assert.Equal(t, typeResult, opts.postType)
	}
}
