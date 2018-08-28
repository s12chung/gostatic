package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/s12chung/gostatic/go/blueprint"
	"io/ioutil"
)

const goStaticDownloadUrl = "https://codeload.github.com/s12chung/gostatic/zip/master"
const goStaticZipFilename = "gostatic-master"

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
		err := initProject(projectName)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
	},
}

func initProject(projectName string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	namespace, err := getNamepace(pwd, projectName)
	if err != nil {
		return err
	}

	fmt.Print("\n")
	tempDir, err := ioutil.TempDir("", "gostatic")
	if err != nil {
		return err
	}
	defer func() { os.RemoveAll(tempDir) }()
	zipPath := path.Join(tempDir, goStaticZipFilename+".zip")
	fmt.Printf("Downloading gostatic@master from %v.\n", goStaticDownloadUrl)
	err = downloadfile(goStaticDownloadUrl, zipPath)
	if err != nil {
		return err
	}
	fmt.Printf("Downloaded and unzipping in temp dir.\n")
	err = unzip(zipPath, tempDir)
	if err != nil {
		return err
	}
	fmt.Print("\n")

	bp := blueprint.NewBlueprint(path.Join(tempDir, goStaticZipFilename, "blueprint"), pwd, namespace)
	bpMessage, err := bp.Init()
	if err != nil {
		return err
	}

	err = exec.Command("direnv", "version").Run()
	hasDirEnv := true
	if err != nil {
		hasDirEnv = false
		fmt.Println("direnv not installed, please note that Makefile and Docker use the environment variables set in .envrc.")
	}

	err = exec.Command("docker", "-v").Run()
	if err != nil {
		fmt.Println("docker not installed. To install locally, you can see the Dockerfile to view system dependencies.")
	}

	env := "DOCKER_WORKDIR=" + "/go/src/" + namespace + " "
	if hasDirEnv {
		env = ""
	}

	fmt.Print("\n")
	fmt.Print(bpMessage)
	fmt.Print("\n")
	fmt.Printf("Project creation success! Install the project in docker via: `%vmake docker-install`\n", env)
	return nil
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
	} else {
		namespace = fmt.Sprintf("%v/%v", rel, projectName)
		fmt.Println("In $GOPATH/src, defaulting namespace to: " + namespace)
	}
	return namespace, nil
}
