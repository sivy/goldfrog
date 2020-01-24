package webmention

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	webmentionRocks string = "https://webmention.rocks"
)

type discoTest struct {
	sourceUrl   string
	wmUrl       string
	description string
}

var wmDiscoTests = []discoTest{
	{"/test/1", "/test/1/webmention", "HTTP Link header, unquoted rel, relative URL"},
	{"/test/2", "/test/2/webmention", "HTTP Link header, unquoted rel, absolute URL"},
	{"/test/3", "/test/3/webmention", "HTML <link> tag, relative URL"},
	{"/test/4", "/test/4/webmention", "HTML <link> tag, absolute URL"},
	{"/test/5", "/test/5/webmention", "HTML <a> tag, relative URL"},
	{"/test/6", "/test/6/webmention", "HTML <a> tag, absolute URL"},
	{"/test/7", "/test/7/webmention", "HTTP Link header with strange casing"},
	{"/test/8", "/test/8/webmention", "HTTP Link header, quoted rel"},
	{"/test/9", "/test/9/webmention", "Multiple rel values on a <link> tag"},
	{"/test/10", "/test/10/webmention", "Multiple rel values on a Link header"},
	{"/test/11", "/test/11/webmention",
		"Multiple Webmention endpoints advertised: Link, <link>, <a>"},
	{"/test/12", "/test/12/webmention", "Checking for exact match of rel=webmention"},
	{"/test/13", "/test/13/webmention", "False endpoint inside an HTML comment"},
	{"/test/14", "/test/14/webmention", "False endpoint in escaped HTML"},
	{"/test/15", "/test/15", "Webmention href is an empty string"},
	{"/test/16", "/test/16/webmention", "Multiple Webmention endpoints advertised: <a>, <link>"},
	{"/test/17", "/test/17/webmention", "Multiple Webmention endpoints advertised: <link>, <a>"},
	{"/test/18", "/test/18/webmention", "Multiple HTTP Link headers"},
	{"/test/19", "/test/19/webmention", "Single HTTP Link header with multiple values"},
	{"/test/20", "/test/20/webmention", "Link tag with no href attribute"},
	{"/test/21", "/test/21/webmention?query=yes", "Webmention endpoint has query string parameters"},
	{"/test/22", "/test/22/webmention", "Webmention endpoint is relative to the path"},
}

func TestDiscovery(t *testing.T) {

	for i, dt := range wmDiscoTests {
		testNum := i + 1
		url := webmentionRocks + dt.sourceUrl
		client := NewWebMentionClient()

		expected := webmentionRocks + dt.wmUrl
		endpoint, err := client.EndpointDiscovery(url)
		if err != nil {
			assert.Fail(t, fmt.Sprintf("[%d] %s: %v", testNum, dt.description, err))
		}
		assert.Equal(t, expected, endpoint, dt.description)
	}

	// #23
	url := webmentionRocks + "/test/23/page"
	client := NewWebMentionClient()
	endpoint, err := client.EndpointDiscovery(url)
	message := "Webmention target is a redirect and the endpoint is relative"

	if err != nil {
		assert.Fail(t, fmt.Sprintf("%s: %v", message, err))
	}
	assert.NotEmpty(t, endpoint, message)

}

var testHTML = `
<!DOCTYPE html>

<html>

<head>
    <title>webmention test</title>
</head>

<body>
    <ol>
        <li class="h-entry"><a href="#p1" class="u-url u-uid" name="p1">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/1"> Test 1</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p2" class="u-url u-uid" name="p2">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/2"> Test 2</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p3" class="u-url u-uid" name="p3">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/3"> Test 3</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p4" class="u-url u-uid" name="p4">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/4"> Test 4</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p5" class="u-url u-uid" name="p5">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/5"> Test 5</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p6" class="u-url u-uid" name="p6">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/6"> Test 6</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p7" class="u-url u-uid" name="p7">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/7"> Test 7</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p8" class="u-url u-uid" name="p8">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/8"> Test 8</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p9" class="u-url u-uid" name="p9">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/9"> Test 9</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p10" class="u-url u-uid" name="p10">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/10"> Test 10</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p11" class="u-url u-uid" name="p11">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/11"> Test 11</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p12" class="u-url u-uid" name="p12">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/12"> Test 12</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p13" class="u-url u-uid" name="p13">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/13"> Test 13</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p14" class="u-url u-uid" name="p14">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/14"> Test 14</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p15" class="u-url u-uid" name="p15">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/15"> Test 15</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p16" class="u-url u-uid" name="p16">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/16"> Test 16</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p17" class="u-url u-uid" name="p17">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/17"> Test 17</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p18" class="u-url u-uid" name="p18">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/18"> Test 18</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p19" class="u-url u-uid" name="p19">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/19"> Test 19</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p20" class="u-url u-uid" name="p20">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/20"> Test 20</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p21" class="u-url u-uid" name="p21">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/21"> Test 21</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p22" class="u-url u-uid" name="p22">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/22"> Test 22</a> by <span class="p-author">Steve Ivy</span></li>
        <li class="h-entry"><a href="#p23" class="u-url u-uid" name="p23">#</a>goldfrog Webmention <a
                href="https://webmention.rocks/test/23/page"> Test 23</a> by <span class="p-author">Steve Ivy</span>
        </li>
    </ol>

</body>

</html>
`

func TestSendMentions(t *testing.T) {

	client := NewWebMentionClient()
	assert.NotNil(t, client)

	links, err := client.FindLinks(testHTML)
	assert.NotNil(t, links)

	if err != nil {
		assert.Fail(t, fmt.Sprintf("%v", err))
	}

	client.SendWebMentions("http://monkinetic.blog/static/webmention.html", links)
}
