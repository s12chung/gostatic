package test

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

// PrintLogEntries prints all the log entries of the hook
func PrintLogEntries(t *testing.T, hook *logTest.Hook) {
	for _, entry := range hook.AllEntries() {
		s, err := entry.String()
		AssertError(t, err, "entry.String")
		t.Log(strings.TrimSpace(s))
	}
}

// SafeLogEntries returns true if all log entries of hook are "safe" (not warnings or more dangerous)
func SafeLogEntries(hook *logTest.Hook) bool {
	for _, entry := range hook.AllEntries() {
		if entry.Level <= logrus.WarnLevel {
			return false
		}
	}
	return true
}

// LogEntryLevels returns an array of all the log levels of the entries
func LogEntryLevels(hook *logTest.Hook) []logrus.Level {
	var levels []logrus.Level
	for _, entry := range hook.AllEntries() {
		levels = append(levels, entry.Level)
	}
	return levels
}
