package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	exitCodeUnsuccessful = 111
	exitCodeFailed       = -1
)

var envNameRegex = regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*$`)

type environment map[string]string

func main() {
	// Register root command in cobra
	var rootCmd = &cobra.Command{
		Use:                "genvdir dir prog...",
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			env := make(environment)
			env.loadEnv()
			if err := env.loadEnvDir(args[0]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exitCodeUnsuccessful)
			}

			binary := whichCmd(args[1], env)
			if err := runCommand(binary, args[1:], env.toArray()); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exitCodeFailed)
			}
		},
	}
	rootCmd.Execute()
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

func (env environment) loadEnv() {
	for _, pair := range os.Environ() {
		assoc := strings.SplitN(pair, "=", 2)
		key := strings.Join(assoc[:1], "")
		value := strings.Join(assoc[1:], "=")
		env[key] = value
	}
}

func (env environment) loadEnvDir(directory string) error {
	// Iterate over each file in directory
	contents, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
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
			if fileLocation, err = os.Readlink(fileLocation); err != nil {
				return fmt.Errorf("Error: Cannot read symlink: \"%s\"\n", fileLocation)
			}

			if !filepath.IsAbs(fileLocation) {
				fileLocation = path.Join(directory, fileLocation)
			}

			if file, err = os.Stat(fileLocation); err != nil {
				return fmt.Errorf("Error: Failed to call os.Stat of %s\n", fileLocation)
			}

			if file.IsDir() {
				continue
			}
		}

		// Read content
		fileData, err := ioutil.ReadFile(fileLocation)
		if err != nil {
			return err
		}

		// Sanitize
		fileString := string(fileData)
		if containsNullChar(fileString) {
			return fmt.Errorf("Error: %s contains a null character\n", fileName)
		}
		fileString = trim(fileString)
		if len(fileString) == 0 {
			delete(env, fileName)
			continue
		}
		env[fileName] = fileString
	}
	return nil
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

func whichCmd(
	name string,
	envs environment,
) string {
	if paths := envs["PATH"]; len(paths) > 0 {
		for _, path := range strings.Split(paths, ":") {
			binary := fmt.Sprintf("%s/%s", path, name)
			if _, err := os.Stat(binary); !os.IsNotExist(err) {
				return binary
			}
		}
	}
	return name
}

func runCommand(
	name string,
	args []string,
	envs []string,
) error {
	if err := syscall.Exec(name, args, envs); err != nil {
		return fmt.Errorf("Error: %s\n", err.Error())
	}
	return nil
}
