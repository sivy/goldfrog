package webmention

import (
	"errors"
	"strings"
	"time"

	"github.com/andyleap/microformats"
)

type State int

const (
	PENDING State = iota
	APPROVED
	REJECTED
)

var ApprovalLabels = map[State]string{
	PENDING:  "Pending",
	APPROVED: "Approved",
	REJECTED: "Rejected",
}

type WebMentionAuthor struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Photo string `json:"photo"`
	Note  string `json:"note"`
}

/*
The webmention process is still a work-in-process. Tantek and
others have been brainstorming what to store.

https://indieweb.org/Webmention-brainstorming#storage

  source (url)
  target (url)
  datetime received
  validation state (pass/fall/pending/etc)
  datetime validated
  source blob (whole html payload or just a section of the html payload)

Then you should process that blob from the source. If you can learn additional information about the source's post type, you should store that (rsvp, photo, note, etc). Then process the blob for type specific properties.

Optionally store:

  source's post type
  source's post specific properties (either inline or in a referenced store (e.g., a photos table in a database)
*/
type WebMention struct {
	Type          string           `json:"type"`
	Target        string           `json:"target"`
	SourceHTML    string           `json:"sourceHtml"`
	Author        WebMentionAuthor `json:"author"`
	Published     time.Time        `json:"published"`
	ContentText   string           `json:"content"`
	ContentHTML   string           `json:"contentHTML"`
	ApprovedState State            `json:"approved"`
}

func (wm *WebMention) Approved() string {
	return ApprovalLabels[wm.ApprovedState]
}

func (wm *WebMention) AsHEntry() (*microformats.MicroFormat, error) {
	hentry := getSourceHEntry(wm.Target, strings.NewReader(wm.SourceHTML))
	if hentry == nil {
		return hentry, errors.New("Could not parse WebMention source")
	}
	return hentry, nil
}

func (wm *WebMention) AsYAML

func NewWebMention() WebMention {
	return WebMention{
		Type:          "mention",
		ApprovedState: PENDING,
	}
}
