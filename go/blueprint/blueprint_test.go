package blueprint

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/s12chung/fastwalk"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/gostatic/go/lib/utils"
	"github.com/s12chung/gostatic/go/test"
)

var updateFixturesPtr = test.UpdateFixtureFlag()

func defaultBlueprint(t *testing.T) (*Blueprint, func()) {
	dir, clean := test.SandboxDir(t, "")
	return NewBlueprint(path.Join(test.FixturePath, "blueprint"), dir, "github.com/s12chung/testproject"), clean
}

func dirFilenames(dirPath string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	filenames := make([]string, len(fileInfos))
	for i, fileInfo := range fileInfos {
		filenames[i] = fileInfo.Name()
	}
	sort.Strings(filenames)
	return filenames, nil
}

func setupGitIgnore(t *testing.T, blueprint *Blueprint) (func(), error) {
	gitIgnoreFilePath := path.Join(blueprint.srcDir, gitIgnoreFilename)
	testGitIgnorePath := path.Join(blueprint.srcDir, gitIgnoreFilename+".test")

	err := utils.CopyFile(testGitIgnorePath, gitIgnoreFilePath)
	clean := func() {
		if e := os.Remove(gitIgnoreFilePath); e != nil {
			t.Error(e)
		}
	}
	if err != nil {
		return clean, err
	}

	err = os.Remove(testGitIgnorePath)
	if err != nil {
		t.Error(err)
	}
	return func() {
		if e := utils.CopyFile(gitIgnoreFilePath, testGitIgnorePath); e != nil {
			t.Error(e)
		}
		clean()
	}, nil
}

func TestBlueprint_InitProject(t *testing.T) {
	blueprint, clean := defaultBlueprint(t)
	defer clean()

	cleanGitIgnore, err := setupGitIgnore(t, blueprint)
	defer cleanGitIgnore()
	if err != nil {
		t.Error(err)
		return
	}

	msg, err := blueprint.NewProject()
	if err != nil {
		t.Error(err)
		return
	}
	if msg == "" {
		t.Error("Did not expect blueprint.NewProject() msg to be empty")
	}

	err = os.Rename(path.Join(blueprint.ProjectDir(), gitIgnoreFilename), path.Join(blueprint.ProjectDir(), gitIgnoreFilename+".test"))
	if err != nil {
		t.Error(err)
		return
	}

	expDir := path.Join(test.FixturePath, "exp")
	err = fastwalk.Walk(blueprint.ProjectDir(), walkAndCompareF(t, func(p string) string {
		return path.Join(expDir, strings.TrimPrefix(p, blueprint.ProjectDir()))
	}))
	if err != nil {
		t.Error(err)
	}

	msg, err = blueprint.NewProject()
	if err != nil {
		t.Error(err)
		return
	}
	if msg != "" {
		t.Error("Expect blueprint.NewProject() msg to be empty because .example real files already exist")
	}
}

func walkAndCompareF(t *testing.T, destPathF func(p string) string) func(p string, typ os.FileMode) error {
	return func(p string, typ os.FileMode) error {
		destPath := destPathF(p)
		if *updateFixturesPtr {
			if typ.IsDir() {
				return utils.MkdirAll(destPath)
			}
			return utils.CopyFile(p, destPath)
		}

		if typ.IsDir() {
			projectFilenames, err := dirFilenames(p)
			if err != nil {
				return err
			}
			expFileNames, err := dirFilenames(destPath)
			if err != nil {
				return err
			}
			if !cmp.Equal(projectFilenames, expFileNames) {
				t.Error(test.DiffString("projectFilenames for "+p, projectFilenames, expFileNames, cmp.Diff(projectFilenames, expFileNames)))
			}
			return nil
		}
		projectFile, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		expFile, err := ioutil.ReadFile(destPath)
		if err != nil {
			return err
		}

		if !bytes.Equal(projectFile, expFile) {
			t.Errorf("%v and %v are diff", p, destPath)
		}
		return nil
	}
}
