package webmention

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebMention(t *testing.T) {
	wm := NewWebMention()
	assert.Equal(t, wm.Type, "mention")
	assert.Equal(t, wm.ApprovedState, PENDING)
}

func TestWebMentionState(t *testing.T) {
	wm := NewWebMention()
	assert.Equal(t, wm.Approved(), ApprovalLabels[wm.ApprovedState])

	wm.ApprovedState = APPROVED
	assert.Equal(t, wm.Approved(), ApprovalLabels[wm.ApprovedState])

	wm.ApprovedState = REJECTED
	assert.Equal(t, wm.Approved(), ApprovalLabels[wm.ApprovedState])
}

func TestWebMentionSOurceToHEntry(t *testing.T) {
	source := `
	<div class="h-entry">
		<div class="u-in-reply-to h-cite">
			<a class="p-name u-url" href="http://example.com/post">Example Post</a>
		</div>
	</div>
	`

	wm := NewWebMention()
	wm.Target = "http://example.com/post"
	wm.SourceHTML = source

	hentry := getSourceHEntry("http://example.com/post", strings.NewReader(source))
	wmHEntry, err := wm.AsHEntry()
	assert.Nil(t, err)

	assert.Equal(t, wmHEntry, hentry)

	wm.SourceHTML = `<div></div>`
	wmHEntry, err = wm.AsHEntry()
	assert.NotNil(t, err)

}
