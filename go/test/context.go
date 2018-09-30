package test

import (
	"fmt"
	"sort"
	"strings"
)

// ContextFields is the fields of the context
type ContextFields map[string]interface{}

// Context represents a context of a test loop
type Context struct {
	fields       ContextFields
	fieldsString string
}

// NewContext returns a new Context
func NewContext() *Context {
	return &Context{}
}

// SetFields sets the fields of Context
func (context *Context) SetFields(fields ContextFields) *Context {
	context.fields = fields
	context.fieldsString = ""
	return context
}

func (context *Context) makeFieldsString() string {
	fieldStrings := make([]string, len(context.fields))
	i := 0
	for k, v := range context.fields {
		fieldStrings[i] = fmt.Sprintf("%v=%v", k, v)
		i++
	}
	sort.Strings(fieldStrings)
	return strings.Join(fieldStrings, " ")
}

// FieldsString returns the fields of the context as a sorted string of key1=value1 key2=value2 ...
func (context *Context) FieldsString() string {
	if context.fieldsString == "" {
		context.fieldsString = context.makeFieldsString()
	}
	return context.fieldsString
}

// String returns the string representation of i, prefixed with the FieldsString()
func (context *Context) String(i interface{}) string {
	return context.Stringf("%v", i)
}

// Stringf is a formatted version of String()
func (context *Context) Stringf(format string, args ...interface{}) string {
	return strings.Join([]string{context.FieldsString(), fmt.Sprintf(format, args...)}, " - ")
}

// GotExpString is String() for assertions
func (context *Context) GotExpString(label string, got, exp interface{}) string {
	return context.String(AssertLabelString(label, got, exp))
}

// DiffString is String() for diffs
func (context *Context) DiffString(label string, got, exp, diff interface{}) string {
	return context.Stringf(DiffString(label, got, exp, diff))
}
