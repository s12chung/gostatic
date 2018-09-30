package webpack

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"sync"
)

const manifestPath = "manifest.json"

// Manifest represents a Manifest file
type Manifest struct {
	generatedPath    string
	assetsFolder     string
	manifestMap      map[string]string
	manifestMapMutex *sync.RWMutex
	log              logrus.FieldLogger
}

// NewManifest returns a new instance of Manifest
func NewManifest(generatedPath, assetsFolder string, log logrus.FieldLogger) *Manifest {
	return &Manifest{
		generatedPath,
		assetsFolder,
		map[string]string{},
		&sync.RWMutex{},
		log,
	}
}

// ManifestURL returns the manifest URL of the file (so it returns hashed file paths that exist), given a file path key.
func (w *Manifest) ManifestURL(key string) string {
	return w.assetsFolder + "/" + w.manifestValue(key)
}

func (w *Manifest) manifestValue(key string) string {
	w.manifestMapMutex.Lock()
	if len(w.manifestMap) == 0 {
		err := w.readManifest()
		if err != nil {
			w.log.Errorf("readManifest error: %v", err)
			return key
		}
	}
	w.manifestMapMutex.Unlock()

	w.manifestMapMutex.RLock()
	value := w.manifestMap[key]
	w.manifestMapMutex.RUnlock()

	if value == "" {
		w.log.Errorf("webpack manifestValue not found for key: %v", key)
		return key
	}
	return value
}

func (w *Manifest) readManifest() error {
	bytes, err := ioutil.ReadFile(filepath.Join(w.generatedPath, w.assetsFolder, manifestPath))
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &w.manifestMap)
}
