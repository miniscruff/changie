package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
)

// TimeNow is a func type for time.Now
type TimeNow func() time.Time

type EditorRunner interface {
	Run() error
}

var (
	ErrBadVersionOrPart      = errors.New("part string is not a supported version or version increment")
	ErrMissingAutoLevel      = errors.New("kind config missing auto level value for auto bumping")
	ErrNoChangesFoundForAuto = errors.New("no unreleased changes found for automatic bumping")
	ErrKindNotFound          = errors.New("kind not found but configuration expects one")
)

var (
	bom = []byte{0xef, 0xbb, 0xbf}
)

func AppendFile(rootFile io.Writer, path string) error {
	otherFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	_, err = io.Copy(rootFile, otherFile)

	return err
}

func ValidBumpLevel(level string) bool {
	return level == MajorLevel ||
		level == MinorLevel ||
		level == PatchLevel ||
		level == AutoLevel
}

func WriteNewlines(writer io.Writer, lines int) error {
	if lines == 0 {
		return nil
	}

	_, err := writer.Write([]byte(strings.Repeat("\n", lines)))

	return err
}

func EnvVarMap(environ []string) map[string]string {
	ret := make(map[string]string)

	for _, env := range environ {
		split := strings.SplitN(env, "=", 2)
		ret[split[0]] = split[1]
	}

	return ret
}

func LoadEnvVars(config *Config, envs []string) map[string]string {
	ret := make(map[string]string)
	if config.EnvPrefix == "" {
		return ret
	}

	for k, v := range EnvVarMap(envs) {
		key, found := strings.CutPrefix(k, config.EnvPrefix)
		if found {
			ret[key] = v
		}
	}

	return ret
}

func FileExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		return !fi.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// createTempFile will create a new temporary file, writing a BOM header if we need to.
// It will return the path to that file or an error.
func createTempFile(runtime string, ext string) (string, error) {
	file, err := os.CreateTemp("", "changie-body-txt-*."+ext)
	if err != nil {
		return "", err
	}

	defer file.Close()

	// The reason why we do this is because notepad.exe on Windows determines the
	// encoding of an "empty" text file by the locale, for example, GBK in China,
	// while golang string only handles utf8 well. However, a text file with utf8
	// BOM header is not considered "empty" on Windows, and the encoding will then
	// be determined utf8 by notepad.exe, instead of GBK or other encodings.
	// This could be enhanced in the future by doing this only when a non-utf8
	// locale is in use, and possibly doing that for any OS, not just windows.
	if runtime == "windows" {
		if _, err = file.Write(bom); err != nil {
			return "", err
		}
	}

	return file.Name(), nil
}

// BuildCommand will create an exec command to run our editor.
func BuildCommand(editorFilePath string) (EditorRunner, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, errors.New("'EDITOR' env variable not set")
	}

	args, err := shellquote.Split(editor)
	if err != nil {
		return nil, err
	}

	args = append(args, editorFilePath)

	// #nosec G204
	cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)

	// Set the stdin and stdout of the command to the current process's stdin and stdout
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd, nil
}

// getBodyTextWithEditor will run the provided editor runner and read the final file.
func getBodyTextWithEditor(runner EditorRunner, editorFile string) (string, error) {
	if err := runner.Run(); err != nil {
		return "", fmt.Errorf("opening the editor: %w", err)
	}

	buf, err := os.ReadFile(editorFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes.TrimPrefix(buf, bom))), nil
}
