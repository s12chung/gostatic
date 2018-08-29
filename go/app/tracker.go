package app

import "fmt"

// A struct to track all the routes/Urls of the static site,
// so that App knows to generate them
type Tracker struct {
	allUrls       func() ([]string, error)
	dependentUrls map[string]bool
}

func NewTracker(allUrls func() ([]string, error)) *Tracker {
	return &Tracker{allUrls, map[string]bool{}}
}

// When generating the static website files, App works in 2 stages.
// First, the IndependentUrls() routes are run, then the DependentUrls() are run.
// This adds a dependent url to move this route into the second stage.
func (tracker *Tracker) AddDependentUrl(url string) {
	tracker.dependentUrls[url] = true
}

// IndependentUrls = AllUrls - DependentUrls (see AddDependentUrl)
func (tracker *Tracker) IndependentUrls() ([]string, error) {
	allUrls, err := tracker.allUrls()
	if err != nil {
		return nil, err
	}

	independentUrlsLen := len(allUrls) - len(tracker.dependentUrls)
	independentUrls := make([]string, independentUrlsLen)
	i := 0
	for _, url := range allUrls {
		if !tracker.dependentUrls[url] {
			if i == independentUrlsLen {
				return nil, fmt.Errorf("there are dependentUrls that are not in allUrls")
			}
			independentUrls[i] = url
			i++
		}
	}
	return independentUrls, nil
}

// DependentUrls returns a slice of urls provided by AddDependentUrl
func (tracker *Tracker) DependentUrls() []string {
	urls := make([]string, len(tracker.dependentUrls))
	i := 0
	for url := range tracker.dependentUrls {
		urls[i] = url
		i += 1
	}
	return urls
}
