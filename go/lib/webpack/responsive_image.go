package webpack

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/utils"
)

// ResponsiveImage represents a ResponsiveImage
type ResponsiveImage struct {
	Src    string `json:"src"`
	SrcSet string `json:"srcSet"`
}

var spacesRegex = regexp.MustCompile(`\s+`)

// PrependSrcPath prepends the given prefix to the Src and SrcSet of the ResponsiveImage
func (r *ResponsiveImage) PrependSrcPath(prefix string, log logrus.FieldLogger) {
	r.Src = prependSrcPath(prefix, r.Src)
	if r.SrcSet == "" {
		return
	}

	var newSrcSet []string
	for _, srcWidth := range strings.Split(r.SrcSet, ",") {
		srcWidthSplit := spacesRegex.Split(strings.Trim(srcWidth, " "), -1)
		if len(srcWidthSplit) != 2 {
			log.Warn("skipping, srcSet is not formatted correctly with '%v' for img src='%v'", srcWidth, r.Src)
			continue
		}
		newSrcSet = append(newSrcSet, fmt.Sprintf("%v %v", prependSrcPath(prefix, srcWidthSplit[0]), srcWidthSplit[1]))
	}

	r.SrcSet = strings.Join(newSrcSet, ", ")
}

// HTMLAttrs returns the HTML attributes of the ResponsiveImage for the <img> tag
func (r *ResponsiveImage) HTMLAttrs() string {
	var htmlAttrs []string
	if r.Src != "" {
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`src="%v"`, r.Src))
	}
	if r.SrcSet != "" {
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`srcset="%v"`, r.SrcSet))
	}
	return strings.Join(htmlAttrs, " ")
}

func prependSrcPath(prefix, src string) string {
	if src == "" {
		return ""
	}

	prefix = utils.CleanFilePath(prefix)
	if prefix == "" {
		return src
	}
	return path.Join(prefix, src)
}
