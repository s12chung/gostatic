package atom

import (
	"strings"
	"time"
)

// EntryLimit is the recommended entry limit for the Atom feed
const EntryLimit = 100

type HTMLEntry struct {
	ID          string
	Title       string
	Updated     time.Time
	HTMLContent string
	Summary     string
	Published   time.Time
}

func (htmlEntry *HTMLEntry) ToEntry(a *Renderer) *Entry {
	return &Entry{
		ID:      strings.Join([]string{a.Settings.Host, htmlEntry.ID, htmlEntry.Published.Format("2006-01-02")}, ":"),
		Title:   htmlEntry.Title,
		Updated: htmlEntry.Updated,

		Author:  a.Author(),
		Content: &EntryContent{Content: htmlEntry.HTMLContent, Type: "html"},
		Summary: htmlEntry.Summary,
		Link:    a.AlternateLink(htmlEntry.ID),

		Published: htmlEntry.Published,
	}
}

type HTMLRenderer struct {
	Settings *Settings
}

func NewHTMLRenderer(settings *Settings) *HTMLRenderer {
	return &HTMLRenderer{settings}
}

func (renderer *HTMLRenderer) Render(feedName, selfURL, logoURL string, htmlEntries []*HTMLEntry) ([]byte, error) {
	atomRenderer := NewRenderer(renderer.Settings)
	feed := HTMLEntriesToFeed(atomRenderer, feedName, selfURL, logoURL, htmlEntries)
	return feed.Marhshall()
}

func HTMLEntriesToFeed(atomRenderer *Renderer, feedName, selfURL, logoURL string, htmlEntries []*HTMLEntry) *Feed {
	entries := make([]*Entry, len(htmlEntries))
	for i, htmlEntry := range htmlEntries {
		entries[i] = htmlEntry.ToEntry(atomRenderer)
	}

	lastUpdated := time.Now()
	if len(htmlEntries) >= 1 {
		lastUpdated = htmlEntries[0].Updated
	}
	feed := atomRenderer.NewFeed(feedName, lastUpdated, selfURL, logoURL)
	feed.Entries = entries
	return feed
}
