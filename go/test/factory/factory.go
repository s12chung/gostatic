/*
Package factory is a set of factory funnctions
*/
package factory

import (
	"math/rand"
	"time"
)

// RandSeed sets the rand.Seed
func RandSeed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TimeDiff finds the time difference between the callack
func TimeDiff(callback func()) time.Duration {
	start := time.Now()
	callback()
	return time.Since(start)
}

// Time returns a standard time to test with
func Time(i int) time.Time {
	return time.Date(2018, 1, i, i, i, i, i, time.UTC)
}
