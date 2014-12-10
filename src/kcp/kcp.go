package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"parseargs"
	"sort"
	"strings"
)

const (
	versionNumber   = "0.21"
	author          = "B. VAUDOUR"
	description     = "Tool in command-line for KaOS Community Packages"
	longDescription = `Provides a tool to make the use of KaOS Community Packages.

With this tool, you can search, get and install a package from KaOS Community Packages.`
	defaultEditor  = "vim"
	searchHead     = "application/vnd.github.v3.text-match+json"
	searchHeadType = "Accept"
	searchBase     = "https://api.github.com/search/repositories?q=%v+user:KaOS-Community-Packages+fork:true"
	urlBase        = "https://github.com/KaOS-Community-Packages/%v.git"
	urlPkgbuild    = "https://raw.githubusercontent.com/KaOS-Community-Packages/%v/master/PKGBUILD"
	kcp_lock       = "kcp.lock"
)

type searcher struct {
	name         string
	localversion string
	kcpversion   string
	description  string
	stars        int64
}

func news(e map[string]interface{}) *searcher {
	s := new(searcher)
	s.update(e)
	return s
}
func (s *searcher) update(e map[string]interface{}) {
	s.name = fmt.Sprintf("%s", e["name"])
	s.description = fmt.Sprintf("%s", e["description"])
	s.stars = int64(e["stargazers_count"].(float64))
}
func (s *searcher) String() string {
	var out string
	switch {
	case s.kcpversion == "":
		if s.localversion == "" {
			out = fmt.Sprintf("\033[1m%v\033[m\033 \033[1;34m(%v)\033[m\n", s.name, s.stars)
		} else {
			out = fmt.Sprintf("\033[1m%v\033[m \033[1;36m[installed: %v]\033[m \033[1;34m(%v)\033[m\n", s.name, s.localversion, s.stars)
		}
	case s.localversion == "":
		out = fmt.Sprintf("\033[1m%v\033[m \033[1;32m%v\033[m\033[1;36m\033[m \033[1;34m(%v)\033[m\n", s.name, s.kcpversion, s.stars)
	case s.localversion == s.kcpversion:
		out = fmt.Sprintf("\033[1m%v\033[m \033[1;32m%v\033[m\033[1;36m [installed]\033[m \033[1;34m(%v)\033[m\n", s.name, s.kcpversion, s.stars)
	default:
		out = fmt.Sprintf("\033[1m%v\033[m \033[1;32m%v\033[m\033[1;36m [installed: %v]\033[m \033[1;34m(%v)\033[m\n", s.name, s.kcpversion, s.localversion, s.stars)
	}
	out = fmt.Sprintf("%s\t%s", out, s.description)
	return out
}

type searchers []*searcher

func (s searchers) Len() int      { return len(s) }
func (s searchers) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s searchers) Less(i, j int) bool {
	if s[i].stars != s[j].stars {
		return s[i].stars > s[j].stars
	}
	return s[i].name <= s[j].name
}

var editor string
var tmpDir string
var fGet, fInstall, fSearch *string
var fHelp, fVersion, fFast, fDeps, fOutdated, fStars *bool
var fList, fListStarred *bool
var p *parseargs.Parser

func init() {
	editor = os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}
	tmpDir = os.TempDir()

	p = parseargs.New(description, versionNumber)
	p.Set(parseargs.SYNOPSIS, "[OPTIONS] [APP]")
	p.Set(parseargs.AUTHOR, author)
	p.Set(parseargs.LONGDESCRIPTION, longDescription)
	g := p.InsertGroup()
	fHelp = p.Bool("-h", "--help", "print this help")
	fVersion = g.Bool("-v", "--version", "print version")
	fSearch = g.String("-s", "--search", "search an app in KCP", "APP", "")
	fFast = p.Bool("", "--fast", "in conjonction with --search, don't print version")
	fStars = p.Bool("", "--stars", "in conjonction with --search, sort by stars")
	fGet = g.String("-g", "--get", "get needed files to build app", "APP", "")
	fInstall = g.String("-i", "--install", "install an app from KCP", "APP", "")
	fDeps = p.Bool("", "--asdeps", "in conjonction with --install, install as a dependence")
	fList = p.Bool("", "--list-all", "list all packages present in repo")
	fListStarred = p.Bool("", "--list-starred", "list all starred packages sorted")
	fOutdated = p.Bool("-o", "--outdated", "display outdated packages")
	p.SetHidden("--list-all")
	p.SetHidden("--list-starred")
	p.Link("--asdeps", "-i")
	p.Link("--fast", "-s")
	p.Link("--stars", "-s")
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

func search(app string) searchers {
	out := make(searchers, 0)
	search := fmt.Sprintf(searchBase, app)
	var f interface{}
	if err := json.Unmarshal(launchRequest(search, true), &f); err != nil {
		return out
	}
	j := f.(map[string]interface{})["items"]
	if j == nil {
		return out
	}
	result := j.([]interface{})
	for _, a := range result {
		e := a.(map[string]interface{})
		out = append(out, news(e))
	}
	return out
}

func searchPackage(app string, fast bool, sortByStars bool) {
	pkgs := search(app)
	if fast {
		for _, p := range pkgs {
			p.localversion = checkInstalled(fmt.Sprintf("%v", p.name))
		}
	} else {
		for _, p := range pkgs {
			p.localversion = checkInstalled(fmt.Sprintf("%v", p.name))
			p.kcpversion = string(getVersion(launchRequest(fmt.Sprintf(urlPkgbuild, p.name), false)))
		}
	}
	if sortByStars {
		sort.Sort(pkgs)
	}
	for _, p := range pkgs {
		fmt.Println(p)
	}
}

func installPackage(app string, asdeps bool) {
	os.Chdir(tmpDir)
	lck := tmpDir + string(os.PathSeparator) + kcp_lock
	_, e := os.Open(lck)
	if e == nil {
		fmt.Println("\033[1;31mAnother instance is running!\033[m")
		os.Exit(1)
	}
	os.Create(lck)
	wDir := tmpDir + string(os.PathSeparator) + app
	end := func() {
		os.RemoveAll(wDir)
		os.Remove(lck)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		end()
		os.Exit(1)
	}()
	getPackage(app)
	defer end()
	if err := os.Chdir(wDir); err != nil {
		fmt.Println(err)
		end()
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
}

func listpage(p int) bool {
	urll := "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100"
	search := fmt.Sprintf(urll, p)
	var f interface{}
	if err := json.Unmarshal(launchRequest(search, true), &f); err != nil {
		return false
	}
	result := f.([]interface{})
	ok := false
	for _, r := range result {
		ok = true
		e := r.(map[string]interface{})
		app := e["name"]
		fmt.Println(app)
	}
	return ok
}

func listAll() {
	for p := 1; listpage(p); p++ {
	}
}

func listStarred() {
	urll := "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100"
	pkgs := make(searchers, 0)
	for p := 1; ; p++ {
		search := fmt.Sprintf(urll, p)
		var f interface{}
		if err := json.Unmarshal(launchRequest(search, true), &f); err != nil {
			break
		}
		result := f.([]interface{})
		ok := false
		for _, r := range result {
			ok = true
			pkgs = append(pkgs, news(r.(map[string]interface{})))
		}
		if !ok {
			break
		}
	}
	sort.Sort(pkgs)
	for _, p := range pkgs {
		fmt.Println(p)
	}
}

func externInstalled() searchers {
	out := make(searchers, 0)
	cmd := exec.Command("pacman", "-Qm")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, l := range lines {
			o := strings.Fields(l)
			if len(o) == 2 {
				s := new(searcher)
				s.name = o[0]
				s.localversion = o[1]
			}
		}
	}
	return out
}

func displayOutdated() {
	outdated := make(searchers, 0)
	localapps := externInstalled()
	for _, s := range localapps {
		for _, kcppapps := range search(s.name) {
			if kcppapps.name != s.name {
				continue
			}
			s.kcpversion = string(getVersion(launchRequest(fmt.Sprintf(urlPkgbuild, s.name), false)))
			if s.kcpversion == s.localversion {
				s.description = kcppapps.description
				s.stars = kcppapps.stars
				outdated = append(outdated, s)
			}
			break
		}
	}
	for _, p := range outdated {
		fmt.Println(p)
	}
}

func main() {
	checkUser()
	err := p.Parse(os.Args)
	switch {
	case err != nil:
		fmt.Println(err)
		p.PrintHelp()
	case *fList:
		listAll()
	case *fListStarred:
		listStarred()
	case *fHelp:
		p.PrintHelp()
	case *fVersion:
		p.PrintVersion()
	case *fGet != "":
		getPackage(*fGet)
	case *fSearch != "":
		searchPackage(*fSearch, *fFast, *fStars)
	case *fInstall != "":
		installPackage(*fInstall, *fDeps)
	case *fOutdated:
		displayOutdated()
	default:
		p.PrintHelp()
	}
}
