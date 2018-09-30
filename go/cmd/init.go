package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/s12chung/gostatic/go/blueprint"
	"github.com/s12chung/gostatic/go/lib/utils"
)

const goStaticDownloadURL = "https://codeload.github.com/s12chung/gostatic/zip/master"
const goStaticZipFilename = "gostatic-master"

var testOnlyFilePaths = []string{"Gopkg.toml"}

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Start a new project",
	Long:  `Start a new project via using a blueprint from the gostatic repo`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires 1 arg, the Project Name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("New project: %v\n", projectName)
		err := initProject(projectName, test)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
	},
}

func initProject(projectName string, test bool) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	namespace, err := getNamepace(pwd, projectName)
	if err != nil {
		return err
	}

	fmt.Print("\n")

	srcDir := "./blueprint"
	if !test {
		var clean func() error
		srcDir, clean, err = downloadSrc()
		defer func() {
			cerr := clean()
			if err != nil {
				err = cerr
			}
		}()
		if err != nil {
			return err
		}
	}

	bp := blueprint.NewBlueprint(srcDir, pwd, namespace)

	err = confirmSafeOverride(bp.ProjectDir())
	if err != nil {
		return err
	}

	bpMessage, err := bp.InitProject()
	if err != nil {
		return err
	}

	if test {
		err = copyOrOverrideTestOnlyFiles(srcDir, bp.ProjectDir())
		if err != nil {
			return err
		}
	}

	endingMessages(bpMessage)
	return nil
}

func downloadSrc() (srcDir string, clean func() error, err error) {
	clean = func() error { return nil }

	tempDir, err := ioutil.TempDir("", "gostatic")
	if err != nil {
		return
	}
	clean = func() error { return os.RemoveAll(tempDir) }

	zipPath := path.Join(tempDir, goStaticZipFilename+".zip")
	fmt.Printf("Downloading gostatic@master from %v.\n", goStaticDownloadURL)
	err = downloadfile(goStaticDownloadURL, zipPath)
	if err != nil {
		return
	}
	fmt.Printf("Downloaded and unzipping in temp dir.\n")
	err = unzip(zipPath, tempDir)
	if err != nil {
		return
	}
	fmt.Print("\n")
	srcDir = path.Join(tempDir, goStaticZipFilename, "blueprint")
	return
}

func confirmSafeOverride(projectDir string) error {
	_, err := os.Stat(projectDir)
	if !os.IsNotExist(err) {
		fmt.Printf("%v already exists, do you want to replace it's files with init?\n", projectDir)

		var yn string
		yn, err = promptStdIn("Replace? (y/n)")
		if err != nil {
			return err
		}
		yn = strings.ToLower(yn[:1])
		if yn != "y" {
			return fmt.Errorf("response did not start with `y`, aborting")
		}
	}
	return nil
}

func copyOrOverrideTestOnlyFiles(srcDir, projectDir string) error {
	for _, filePath := range testOnlyFilePaths {
		return utils.CopyFile(path.Join(srcDir, filePath), path.Join(projectDir, filePath))
	}
	return nil
}

func endingMessages(bpMessage string) {
	err := exec.Command("direnv", "version").Run()
	dirEnvMessage := ""
	if err != nil {
		dirEnvMessage = " (without direnv, you require DOCKER_WORKDIR ENV variable from .envrc file)"
		fmt.Print("direnv not installed, please note that Makefile and Docker use the environment variables set in .envrc file.\n\n")
	}

	err = exec.Command("docker", "-v").Run() // # #nosec G204
	if err != nil {
		fmt.Print("docker not installed. To install locally, you can see the Dockerfile to view system dependencies.\n\n")
	}

	if bpMessage != "" {
		fmt.Print(bpMessage)
	}
	fmt.Print("\n")
	fmt.Printf("Project creation success! Install the project in docker via: `make docker-install`%v\n", dirEnvMessage)
}

func getNamepace(pwd, projectName string) (string, error) {
	goPathSrc := path.Join(os.Getenv("GOPATH"), "src")
	rel, err := filepath.Rel(goPathSrc, pwd)

	var namespace string
	if err != nil || strings.Index(rel, "..") == 0 {
		fmt.Println("Not in $GOPATH/src. Please indicate a Go namespace.")
		namespaceDefault := "github.com/s12chung/" + projectName
		namespace, err = promptStdIn(fmt.Sprintf("Go namespace (%v)", namespaceDefault))
		if err != nil {
			return "", err
		}
		if namespace == "" {
			namespace = namespaceDefault
		}
	} else {
		namespace = path.Join(rel, projectName)
		fmt.Println("In $GOPATH/src, defaulting namespace to: " + namespace)
	}
	return namespace, nil
}
