package main

/*
 * Depends on goopt: https://github.com/droundy/goopt
 */

import (
	"encoding/json"
	"fmt"
	"goopt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	versionNumber  = "0.10"
	defaultEditor  = "vim"
	searchHead     = "application/vnd.github.v3.text-match+json"
	searchHeadType = "Accept"
	searchBase     = "https://api.github.com/search/repositories?q=%v+user:KaOS-Community-Packages+fork:true"
	urlBase        = "https://github.com/KaOS-Community-Packages/%v.git"
	urlPkgbuild    = "https://raw.githubusercontent.com/KaOS-Community-Packages/%v/master/PKGBUILD"
)

var editor string
var tmpDir string
var fGet, fInstall, fSearch *string
var fHelp, fVersion, fFast, fDeps *bool

func init() {
	editor = os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}
	tmpDir = os.TempDir()

	fHelp = goopt.Flag([]string{"-h", "--help"}, []string{}, "show this help message and exit", "")
	fVersion = goopt.Flag([]string{"-v", "--version"}, []string{}, "print version", "")
	fGet = goopt.String([]string{"-g", "--get"}, "", "get needed files to build app")
	fSearch = goopt.String([]string{"-s", "--search"}, "", "search an app in KCP")
	fFast = goopt.Flag([]string{"--fast"}, []string{}, "search without version", "")
	fInstall = goopt.String([]string{"-i", "--install"}, "", "install an app in KCP")
	fDeps = goopt.Flag([]string{"--asdeps"}, []string{}, "install as a dependence", "")

	goopt.Version = "0.10"
	goopt.Summary = "Tool in command-line for KaOS Community Packages"

}

func printError(msg string) {
	fmt.Printf("\033[1;31m%v\033[m\n", msg)
}

func question(msg string, value bool) bool {
	var defaultVal string
	if value {
		defaultVal = "[Y/n]"
	} else {
		defaultVal = "[y/N]"
	}
	fmt.Printf("\033[1;33m%v %v \033[m", msg, defaultVal)
	var response string
	if _, err := fmt.Scanf("%v", &response); err != nil {
		return value
	}
	if len(response) == 0 {
		return value
	}
	response = strings.ToLower(response)
	if strings.HasPrefix(response, "y") {
		return true
	} else if strings.HasPrefix(response, "n") {
		return false
	}
	return value
}

func edit(filename string) {
	cmd := exec.Command(editor, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func checkUser() {
	if e := os.Geteuid(); e == 0 {
		printError("Don't launch this program as root!")
		os.Exit(1)
	}
}

func editPkgbuild() {
	if question("Do you want to edit PKGBUILD?", true) {
		edit("PKGBUILD")
	}
}

func getPackage(app string) {
	url := fmt.Sprintf(urlBase, app)
	cmd := exec.Command("git", "clone", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func launchRequest(search string, withHeader bool) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", search, nil)
	if err != nil {
		return []byte{}
	}
	if withHeader {
		req.Header.Add(searchHeadType, searchHead)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return b
}

func getVersion(result []byte) string {
	var pkgver, pkgrel string
	l := strings.Split(string(result), "\n")
	for _, e := range l {
		e := strings.TrimSpace(e)
		if len(e) <= 7 {
			continue
		}
		if e[:7] == "pkgver=" {
			pkgver = e[7:]
		} else if e[:7] == "pkgrel=" {
			pkgrel = e[7:]
		}
	}
	if pkgver != "" && pkgrel != "" {
		return pkgver + "-" + pkgrel
	}
	return "<unknown>"
}

func checkInstalled(app string) string {
	var localVersion string
	cmd := exec.Command("pacman", "-Q", app)
	if out, err := cmd.Output(); err == nil {
		o := strings.Fields(string(out))
		if len(o) >= 2 {
			localVersion = o[1]
		}
	}
	return localVersion
}

func searchPackage(app string, fast bool) {
	search := fmt.Sprintf(searchBase, app)
	var f interface{}
	if err := json.Unmarshal(launchRequest(search, true), &f); err != nil {
		return
	}
	j := f.(map[string]interface{})["items"]
	result := j.([]interface{})
	for _, a := range result {
		e := a.(map[string]interface{})
		n, d, s := e["name"], e["description"], e["stargazers_count"]
		i := checkInstalled(fmt.Sprintf("%v", n))
		if fast {
			if i != "" {
				fmt.Printf("\033[1m%v\033[m \033[1;36m[installed: %v]\033[m \033[1;34m(%v)\033[m\n", n, i, s)
			} else {
				fmt.Printf("\033[1m%v\033[m\033 \033[1;34m(%v)\033[m\n", n, s)
			}
		} else {
			v := string(getVersion(launchRequest(fmt.Sprintf(urlPkgbuild, n), false)))
			if i != "" {
				if v == i {
					i = " [installed]"
				} else {
					i = fmt.Sprintf(" [installed: %v]", i)
				}
			}
			fmt.Printf("\033[1m%v\033[m \033[1;32m%v\033[m\033[1;36m%v\033[m \033[1;34m(%v)\033[m\n", n, v, i, s)
		}
		fmt.Println("\t", d)
	}
}

func installPackage(app string, asdeps bool) {
	os.Chdir(tmpDir)
	wDir := tmpDir + string(os.PathSeparator) + app
	getPackage(app)
	if err := os.Chdir(wDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	editPkgbuild()
	var cmd *exec.Cmd
	if asdeps {
		cmd = exec.Command("makepkg", "-si", "--asdeps")
	} else {
		cmd = exec.Command("makepkg", "-si")
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	os.RemoveAll(wDir)
}

func main() {
	checkUser()
	goopt.Parse(nil)
	switch {
	case *fVersion:
		fmt.Println(goopt.Version)
	case *fHelp:
		fmt.Println(goopt.Usage())
	case *fGet != "":
		getPackage(*fGet)
	case *fSearch != "":
		searchPackage(*fSearch, *fFast)
	case *fInstall != "":
		installPackage(*fInstall, *fDeps)
	default:
		fmt.Println(goopt.Usage())
	}
}
