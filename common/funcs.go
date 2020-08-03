package common

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/bvaudour/kcp/color"
	"github.com/leonelquinteros/gotext"
)

//Tr returns the translated string.
func Tr(msg string, vars ...interface{}) string {
	return gotext.Get(msg, vars...)
}

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
	return LaunchCommand(DefaultEditor, f)
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
	defstr, resp := Tr(cDefaultYes), ""
	if !defaultResponse {
		defstr = Tr(cDefaultNo)
	}
	fmt.Print(color.Yellow.Format("%s %s", msg, Tr(defstr)))
	if _, e := fmt.Scanf("%v", &resp); e != nil || len(resp) == 0 {
		return defaultResponse
	}
	yes, no := strings.ToLower(Tr(Yes)), strings.ToLower(Tr(No))
	resp = strings.ToLower(resp)
	switch resp[0] {
	case yes[0]:
		return true
	case no[0]:
		return false
	case Yes[0]:
		return true
	case No[0]:
		return false
	}
	return defaultResponse
}

//PrintError print a red message in the stderr.
func PrintError(e interface{}) {
	fmt.Fprintln(os.Stderr, color.Red.Colorize(e))
}

//PrintWarning print a yellow message in the stderr.
func PrintWarning(e interface{}) {
	fmt.Fprintln(os.Stderr, color.Yellow.Colorize(e))
}

//Now returns the UNIX timestamp from the current time
func Now() int64 {
	return time.Now().UTC().Unix()
}

//StrToTimeStamp converts a formatted string date to a timestamp.
func StrToTimestamp(date string) int64 {
	if date == "" {
		return 0
	}
	utc, _ := time.LoadLocation("")
	d, _ := time.ParseInLocation(time.RFC3339, date, utc)
	return d.Unix()
}

//TimestampTostring convertis an UNIX timestamp to a formatted string.
func TimestampToString(unix int64) string {
	if unix == 0 {
		return ""
	}
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}

//FilExists check if the given file or directory exists on the system.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

//JoinRelative returns the complete path of the file.
//- If the path of the file is absolute, it returns it.
//- If it is relative, it returns the absolute path from the base.
func JoinIfRelative(base, file string) string {
	if len(file) > 0 && file[0] == '/' {
		return file
	}
	return path.Join(base, file)
}
