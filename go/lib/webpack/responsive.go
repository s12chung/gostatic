package webpack

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
)

const responsiveFolder = "responsive"

var responsiveExtensions = map[string]bool{
	".png": true,
	".jpg": true,
}

type Responsive struct {
	generatedPath string
	assetsFolder  string
	log           logrus.FieldLogger
}

func NewResponsive(generatedPath, assetsFolder string, log logrus.FieldLogger) *Responsive {
	return &Responsive{generatedPath, assetsFolder, log}
}

func HasResponsive(originalSrc string) bool {
	_, hasResponsive := responsiveExtensions[filepath.Ext(originalSrc)]
	return hasResponsive
}

func (r *Responsive) GetResponsiveImage(originalSrc string) *ResponsiveImage {
	responsiveImage, err := r.getResponsiveImage(originalSrc)
	if err != nil {
		r.log.Errorf("GetResponsiveImage error: %v", err)
		return nil
	}
	return responsiveImage
}

func (r *Responsive) getResponsiveImage(originalSrc string) (*ResponsiveImage, error) {
	u, err := url.Parse(originalSrc)
	if err != nil {
		return nil, err
	}
	if u.Hostname() != "" {
		return &ResponsiveImage{Src: originalSrc}, nil
	}

	responsiveImage, err := r.readResponsiveImageJSON(originalSrc)
	if err != nil {
		return nil, err
	}
	responsiveImage.PrependSrcPath(r.assetsFolder, r.log)
	return responsiveImage, nil
}

func (r *Responsive) readResponsiveImageJSON(originalSrc string) (*ResponsiveImage, error) {
	filename := fmt.Sprintf("%v.json", path.Base(originalSrc))
	filePath := path.Join(r.generatedPath, r.assetsFolder, path.Dir(originalSrc), responsiveFolder, filename)

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	responsiveImage := &ResponsiveImage{}
	err = json.Unmarshal(bytes, responsiveImage)
	if err != nil {
		return nil, err
	}
	return responsiveImage, nil
}
