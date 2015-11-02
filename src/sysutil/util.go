//Package sysutil provides useful functions to interact with the system.
package sysutil

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	DEFAULT_EDITOR = "vim"
	DEFAULT_YES    = "[Y/n]"
	DEFAULT_NO     = "[y/N]"
	AUTHOR         = "B. VAUDOUR"
	VERSION        = "0.81.2"
	KCP_LOCK       = "kcp.lock"
	KCP_DB         = ".kcp.json"
	LOCALE_DIR     = "/usr/share/locale"
)

//LaunchCommand launches a system command.
func LaunchCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

//GetOuptutCommand returns the redirected output of a system command.
func GetOutputCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

//Edit lets the user edit the given file.
func EditFile(f string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = DEFAULT_EDITOR
	}
	return LaunchCommand(editor, f)
}

//InstalledVersion returns the installed version of a package.
func InstalledVersion(app string) string {
	if b, e := GetOutputCommand("pacman", "-Q", app); e == nil {
		f := strings.Fields(string(b))
		if len(f) >= 2 {
			return f[1]
		}
	}
	return ""
}

//Question displays a question to the output and returns the response given by the user.
func Question(msg string) string {
	fmt.Print(msg + " ")
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return strings.TrimSpace(sc.Text())
}

//QuestionYN displays a question to the output and returns the boolean response given by the user.
func QuestionYN(msg string, defaultResponse bool) bool {
	defstr, resp := DEFAULT_YES, ""
	if !defaultResponse {
		defstr = "[y/N]"
	}
	fmt.Printf("\033[1;33m%s %s \033[m", msg, defstr)
	if _, e := fmt.Scanf("%v", &resp); e != nil || len(resp) == 0 {
		return defaultResponse
	}
	resp = strings.ToLower(resp)
	switch {
	case strings.HasPrefix(resp, "y"):
		return true
	case strings.HasPrefix(resp, "n"):
		return false
	default:
		return defaultResponse
	}
}

//PrintError print a red message in the stderr.
func PrintError(e interface{}) { fmt.Fprintf(os.Stderr, "\033[1;31m%v\033[m\n", e) }

//PrintWarning print a yellow message in the stderr.
func PrintWarning(e interface{}) {
	fmt.Fprintf(os.Stderr, "\033[1;33m%v\033[m\n", e)
}
