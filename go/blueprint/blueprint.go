package blueprint

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/s12chung/fastwalk"

	"github.com/s12chung/gostatic/go/lib/utils"
)

const gitIgnoreFilename = ".gitignore"

var extraGitIgnores = []string{
	".git/*",
}
var ignoredFiles = map[string]bool{
	".DS_Store": true,
}

var exampleRegex = regexp.MustCompile(`\.example(\.[a-zA-Z]{1,8})?$`)

const projectNameString = "blueprint"
const namespaceString = "github.com/s12chung/gostatic/blueprint"
const lockFileExt = ".lock"

type ReplaceFunc func(blueprint *Blueprint, srcPath string) (string, error)

func replaceNamespace(blueprint *Blueprint, s string) (string, error) {
	result := strings.Replace(s, namespaceString, blueprint.namespace, -1)
	return result, nil
}

func replaceProjectName(blueprint *Blueprint, s string) (string, error) {
	result := strings.Replace(s, projectNameString, blueprint.ProjectName(), -1)
	return result, nil
}

var extToFuncs = map[string][]ReplaceFunc{
	".go": {replaceNamespace},
}

var filenameToFuncs = map[string][]ReplaceFunc{
	".example.envrc": {replaceNamespace},
	"Makefile":       {replaceProjectName},
	"package.json":   {replaceProjectName},
}

var replaceFuncsMappings = []struct {
	typeToFunc map[string][]ReplaceFunc
	pathFunc   func(string) string
}{
	{extToFuncs, path.Ext},
	{filenameToFuncs, path.Base},
}

type Blueprint struct {
	srcDir    string
	destDir   string
	namespace string
}

func NewBlueprint(srcDir, destDir, namespace string) *Blueprint {
	return &Blueprint{srcDir, destDir, namespace}
}

func (blueprint *Blueprint) Init() (string, error) {
	ignoreMap, err := blueprint.IgnoreMap()
	if err != nil {
		return "", err
	}

	utils.MkdirAll(blueprint.destPath(""))

	var exampleFiles []string
	err = blueprint.IgnoreWalk(ignoreMap, func(srcPath string, typ os.FileMode) error {
		if srcPath == blueprint.srcDir {
			return nil
		}

		destPath := blueprint.destPath(srcPath)
		if typ.IsDir() {
			return utils.MkdirAll(destPath)
		}
		if path.Ext(srcPath) == lockFileExt {
			return nil
		}

		destPaths := []string{destPath}

		if exampleRegex.MatchString(destPath) {
			realFileDestPath := exampleRegex.ReplaceAllString(destPath, "$1")
			_, err = os.Stat(realFileDestPath)
			if os.IsNotExist(err) {
				exampleFiles = append(exampleFiles, blueprint.destRelativePath(realFileDestPath))
				destPaths = append(destPaths, realFileDestPath)
			}
		}

		for _, m := range replaceFuncsMappings {
			replaceFuncs, got := m.typeToFunc[m.pathFunc(srcPath)]
			if got {
				s, err := blueprint.runReplaceFuncs(srcPath, replaceFuncs)
				if err != nil {
					return err
				}

				bytes := []byte(s)
				return forEachDestPath(destPaths, func(destPath string) error {
					return utils.WriteFile(destPath, bytes)
				})
			}
		}
		return forEachDestPath(destPaths, func(destPath string) error {
			return utils.CopyFile(srcPath, destPath)
		})
	})
	if err != nil {
		return "", err
	}
	if len(exampleFiles) == 0 {
		return "", nil
	}

	messageArray := []string{"Note these files, which are in .gitignore and have .example version:"}
	for _, exampleFile := range exampleFiles {
		messageArray = append(messageArray, "- "+exampleFile)
	}
	messageArray = append(messageArray, "You may need to fill in personal data in them, such as AWS credentials.\n")
	return strings.Join(messageArray, "\n"), nil
}

func (blueprint *Blueprint) ProjectName() string {
	return path.Base(blueprint.namespace)
}

func (blueprint *Blueprint) ProjectDir() string {
	return path.Join(blueprint.destDir, blueprint.ProjectName())
}

func (blueprint *Blueprint) srcRelativePath(srcPath string) string {
	return strings.TrimPrefix(srcPath, blueprint.srcDir+"/")
}

func (blueprint *Blueprint) destRelativePath(srcPath string) string {
	return strings.TrimPrefix(srcPath, path.Join(blueprint.destDir, blueprint.ProjectName())+"/")
}

func (blueprint *Blueprint) destPath(srcPath string) string {
	return path.Join(blueprint.destDir, blueprint.ProjectName(), blueprint.srcRelativePath(srcPath))
}

func (blueprint *Blueprint) runReplaceFuncs(srcPath string, funcs []ReplaceFunc) (string, error) {
	bytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return "", err
	}

	s := string(bytes)
	for _, f := range funcs {
		s, err = f(blueprint, s)
		if err != nil {
			return "", err
		}
	}
	return s, nil
}

func (blueprint *Blueprint) IgnoreWalk(ignoreMap map[string]bool, f func(p string, typ os.FileMode) error) error {
	return fastwalk.Walk(blueprint.srcDir, func(p string, typ os.FileMode) error {
		relativePath := blueprint.srcRelativePath(p)
		if typ.IsDir() {
			ignore := ignoreMap[relativePath+"/*"]
			if ignore {
				return filepath.SkipDir
			}
		} else {
			if ignoreMap[relativePath] || ignoredFiles[path.Base(p)] {
				return nil
			}
		}
		return f(p, typ)
	})
}

func forEachDestPath(destPaths []string, f func(destPath string) error) error {
	for _, destPath := range destPaths {
		err := f(destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (blueprint *Blueprint) IgnoreMap() (map[string]bool, error) {
	ignoreMap, err := blueprint.gitIgnoreMap()
	if err != nil {
		return nil, err
	}

	for _, k := range extraGitIgnores {
		ignoreMap[k] = true
	}
	return ignoreMap, nil

}

func (blueprint *Blueprint) gitIgnoreMap() (map[string]bool, error) {
	bytes, err := ioutil.ReadFile(path.Join(blueprint.srcDir, gitIgnoreFilename))
	if err != nil {
		return nil, err
	}

	ignoreMap := map[string]bool{}
	for _, line := range strings.Split(string(bytes), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ignoreMap[line] = true
	}
	return ignoreMap, err
}
