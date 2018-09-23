package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const exitCodeUnsuccessful = 111

var envNameRegex = regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*$`)

type environment map[string]string

func main() {
	// Register main command in cobra
	var main = &cobra.Command{
		Use:                "envdir dir prog...",
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			env := make(environment)
			loadEnv(env, args[0])
			runCommand(args[1], args[2:], env.toArray())
		},
	}
	main.Execute()
}

func (env environment) toArray() []string {
	result := make([]string, len(env))
	index := 0
	for key, value := range env {
		result[index] = fmt.Sprintf("%s=%s", key, value)
		index++
	}
	return result
}

func (env environment) loadCurrent() {
	for _, pair := range os.Environ() {
		assoc := strings.SplitN(pair, "=", 2)
		key := strings.Join(assoc[:1], "")
		value := strings.Join(assoc[1:], "=")
		env[key] = value
	}
}

func loadEnv(env environment, directory string) {
	env.loadCurrent()

	// Iterate over each file in directory
	contents, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCodeUnsuccessful)
	}
	for _, file := range contents {
		// Skip directory
		if file.IsDir() {
			continue
		}

		// Retrieve file name
		fileName := file.Name()

		// Ensure file name is legit
		if !envNameRegex.MatchString(fileName) {
			continue
		}

		// Ensure file isn't empty
		if file.Size() == 0 {
			delete(env, fileName)
			continue
		}

		fileLocation := path.Join(directory, fileName)

		// Handle symlink
		if file.Mode()&os.ModeSymlink != 0 {
			fileLocation, err = os.Readlink(fileLocation)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Cannot read symlink: \"%s\"\n", fileLocation)
				os.Exit(exitCodeUnsuccessful)
			}

			if !filepath.IsAbs(fileLocation) {
				fileLocation = path.Join(directory, fileLocation)
			}

			file, err = os.Stat(fileLocation)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to call os.Stat of %s\n", fileLocation)
				os.Exit(exitCodeUnsuccessful)
			}

			if file.IsDir() {
				continue
			}
		}

		// Read content
		fileData, err := ioutil.ReadFile(fileLocation)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitCodeUnsuccessful)
		}

		// Sanitize
		fileString := string(fileData)
		if containsNullChar(fileString) {
			fmt.Fprintf(os.Stderr, "Error: %s contains a null character\n", fileName)
			os.Exit(exitCodeUnsuccessful)
		}
		fileString = trim(fileString)
		if len(fileString) == 0 {
			delete(env, fileName)
			continue
		}
		env[fileName] = fileString
	}
}

func containsNullChar(s string) bool {
	i := strings.IndexByte(s, '\x00')
	return i != -1
}

func trim(s string) string {
	if strings.HasSuffix(s, "\r\n") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\r") {
		return s[:len(s)-1]
	}
	return s
}

func runCommand(
	name string,
	args []string,
	envs []string,
) {
	command := exec.Command(name, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = envs
	if err := command.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitCodeUnsuccessful)
		}
	}
	os.Exit(0)
}
