/*
	Struct representations of robots.txt
*/
package robots

import "strings"

const EverythingUserAgent = "*"

// ToFileString returns the robots.txt representation of the UserAgents
func ToFileString(userAgents []*UserAgent) string {
	parts := make([]string, len(userAgents))
	for i, userAgent := range userAgents {
		parts[i] = userAgent.ToFileString()
	}
	return strings.Join(parts, "\n\n")
}

// Representation of a user agent and the paths to ignore
type UserAgent struct {
	name  string
	paths []string
}

func NewUserAgent(name string, paths []string) *UserAgent {
	return &UserAgent{name, paths}
}

// ToFileString returns the robots.txt representation of the UserAgent
func (userAgent *UserAgent) ToFileString() string {
	parts := []string{"User-agent: " + userAgent.name}

	for _, path := range userAgent.paths {
		parts = append(parts, "Disallow: "+path)
	}
	return strings.Join(parts, "\n")
}
