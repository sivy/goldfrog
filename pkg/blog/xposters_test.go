package blog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusMessageFromPost(t *testing.T) {
	longPost := Post{
		Title: "Test",
		Body: `For 2020, I'm writing a new blog app. It's just for myself, a toy to remind me why I love the web. It's called [Goldfrog](https://github.com/sivy/goldfrog), and it sounds a bit like "Go, blog!"

		Why in the hack, in this day and age, would I spend time writing my own #blogging software, when you can't sign up for a VPS _anywhere_ without tripping over offers to help you set up Wordpress, or Ghost, or what have you?

		A few reasons.

		### New Year, New You

		2019 was shite-filled, and due to politics, the tech trashfire, and the friction of blogging through several variations of static, git-powered versions of this site, I simply stopped blogging. I've wanted to, but the effort killed the motivation before I could get some words out.

		So I finally decided to write something myself, that did *just* the things I wanted. #goldfrog is written in Go, because while I will love Python to my dying day, my brain needed a kick in the pants this year, which relates to my next point.

		### The Builder's High

		Rands writes eloquently on [the builder's high](https://randsinrepose.com/archives/the-builders-high/). With family engagements and work over the last few years my hobby coding has dropped to almost nil (None if I were writing Python).

		I needed something to reboot my creative juices, and trying to write something I really wanted, that thought would be quick, in a new language, seemed like a good way to go (I did want it, it wasn't easy, and Go hates me. But I'm learning and that feels great!)
		`,
	}

	content := statusMessageFromPost(
		longPost, 280)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "For 2020")
	assert.True(t, len(content) <= 280)

	content = statusMessageFromPost(
		Post{
			Title: "Title",
			Body:  "foo",
		}, 280)
	assert.NotEmpty(t, content)
	assert.Equal(t, content, "foo")

}

func TestStripHTML(t *testing.T) {
	content := "<p>foo</p>"

	content = stripHTML(content)

	assert.NotContains(t, content, "<p>")
}
