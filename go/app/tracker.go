package app

import "fmt"

// Tracker tracks all the routes/URLs of the static site,
// so that App knows to generate them
type Tracker struct {
	urls          func() []string
	dependentURLs map[string]bool
}

// NewTracker returns a new instance of Tracker
func NewTracker(urls func() []string) *Tracker {
	return &Tracker{urls, map[string]bool{}}
}

// AddDependentURL adds a dependent url, so the url route runs in the second batch.
//
// When generating the static website files (App.Generate), App works in 2 batches.
// First, the IndependentURLs() routes are run, then the DependentURLs() are run.
func (tracker *Tracker) AddDependentURL(url string) {
	if url[:1] != "/" {
		url = "/" + url
	}
	tracker.dependentURLs[url] = true
}

// IndependentURLs = AllURLs - DependentURLs (see AddDependentURL)
func (tracker *Tracker) IndependentURLs() ([]string, error) {
	urls := tracker.urls()
	independentUrlsLen := len(urls) - len(tracker.dependentURLs)
	independentUrls := make([]string, independentUrlsLen)
	i := 0
	for _, url := range urls {
		if !tracker.dependentURLs[url] {
			if i == independentUrlsLen {
				return nil, fmt.Errorf("there are dependentURLs that are not in urls")
			}
			independentUrls[i] = url
			i++
		}
	}
	return independentUrls, nil
}

// DependentURLs returns a slice of urls provided by AddDependentURL
func (tracker *Tracker) DependentURLs() []string {
	urls := make([]string, len(tracker.dependentURLs))
	i := 0
	for url := range tracker.dependentURLs {
		urls[i] = url
		i++
	}
	return urls
}
