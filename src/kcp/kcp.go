package main

import (
	"errors"
	"fmt"
	"gettext"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"parser/pargs"
	"parser/pjson"
	"sort"
	"strings"
)

// Needed keys for requests
const (
	NAME          = "name"
	DESCRIPTION   = "description"
	STARS         = "stargazers_count"
	MESSAGE       = "message"
	DOCUMENTATION = "documentation_url"
)

// Needed URLs for requests
const (
	//HEADERPREVIEW = "application/vnd.github.moondragon-preview+json"
	HEADER       = "application/vnd.github.v3+json"
	HEADERMATCH  = "application/vnd.github.v3.text-match+json"
	SEARCH_ALL   = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100&%s"
	SEARCH_ALLST = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100+stars:>=1&%s"
	SEARCH_APP   = "https://api.github.com/search/repositories?q=%v+user:KaOS-Community-Packages+fork:true&%s"
	URL_REPO     = "https://github.com/KaOS-Community-Packages/%v.git"
	URL_PKGBUILD = "https://raw.githubusercontent.com/KaOS-Community-Packages/%v/master/PKGBUILD"
	APP_ID       = "&client_id=11f5f3d9dab26c7fff24"
	SECRET_ID    = "&client_secret=bb456e9fa4e2d0fe2df9e194974c98c2f9133ff5"
	IDENT        = APP_ID + SECRET_ID
)

// Messages
//   -> Find a way to use i18n files
const (
	MSG_NOPACKAGE       = "No package found"
	MSG_NOROOT          = "Don't launch this program as root!"
	MSG_DIREXISTS       = "Dir %s already exists!"
	MSG_ONLYONEINSTANCE = "Another instance of kcp is running!"
	MSG_INTERRUPT       = "Interrupt by user..."
	MSG_EDIT            = "Do you want to edit PKGBUILD?"
	MSG_UNKNOWN         = "Unknown error!"
)

// Parser's descriptions
const (
	LONGDESCRIPTION = `Provides a tool to make the use of KaOS Community Packages.

With this tool, you can search, get and install a package from KaOS Community Packages.`
	VERSION         = "0.27-dev"
	AUTHOR          = "B. VAUDOUR"
	APP_DESCRIPTION = "Tool in command-line for KaOS Community Packages"
	SYNOPSIS        = "[OPTIONS] [APP]"
	D_HELP          = "Print this help"
	D_VERSION       = "Print version"
	D_LIST          = "Display all packages of KCP"
	D_OUTDATED      = "Display all outdated packages from KCP"
	D_SEARCH        = "Search packages in KCP and display them"
	D_GET           = "Download needed files to build a package"
	D_INSTALL       = "Install a package from KCP"
	D_FAST          = "On display action, don't print KCP version"
	D_SORT          = "On display action, sort packages by stars descending"
	D_ASDEPS        = "On install action, install as a dependence"
	VALUENAME       = "<app>"
)

// Other constants
const (
	KCP_LOCK   = "kcp.lock"
	LOCALE_DIR = "/usr/share/locale"
)

// Package informations extractor
//  - name        : Name of the package
//  - localversion: Version of installed package (if installed)
//  - kcpversion  : Version of PKGBUILD present in KCP
//  - description : Description of the package (extracted from KCP)
//  - stars       : Stars number of the package in KCP
type information struct {
	name         string
	localversion string
	kcpversion   string
	description  string
	stars        int64
}

func (i *information) update(o pjson.Object, keys ...string) {
	for _, k := range keys {
		switch k {
		case NAME:
			if s, e := o.GetString(k); e == nil {
				i.name = s
			}
		case DESCRIPTION:
			if s, e := o.GetString(k); e == nil {
				i.description = s
			}
		case STARS:
			if s, e := o.GetNumber(k); e == nil {
				i.stars = int64(s)
			}
		}
	}
}
func (i *information) updateLocalVersion() {
	out := launchCommandWithResult("pacman", "-Q", i.name)
	spl := strings.Fields(out)
	if len(spl) >= 2 {
		i.localversion = spl[1]
	}
}
func (i *information) updateKcpVersion() bool {
	out := string(launchRequest(false, "", URL_PKGBUILD, i.name))
	if out == "" {
		return false
	}
	pkgver, pkgrel := "", ""
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "pkgver=") {
			pkgver = l[7:]
		} else if strings.HasPrefix(l, "pkgrel=") {
			pkgrel = l[7:]
		}
		if pkgver != "" && pkgrel != "" {
			i.kcpversion = pkgver + "-" + pkgrel
			return true
		}
	}
	i.kcpversion = t("<unknown>")
	return false
}
func (i *information) String() string {
	if i.description == "" {
		return i.name
	}
	out, local, kcp := "", "", ""
	if i.localversion != "" {
		if i.localversion == i.kcpversion {
			local = t(" [installed]")
		} else {
			local = fmt.Sprintf(t(" [installed: %s]"), i.localversion)
		}
	}
	if i.kcpversion != "" {
		kcp = " " + i.kcpversion
	}
	out = fmt.Sprintf("\033[1m%s\033[m\033[1;32m%s\033[m\033[1;36m%s\033[m \033[1;34m(%v)\033[m", i.name, kcp, local, i.stars)
	if i.description != "" {
		out = fmt.Sprintf("%s\n\t%s", out, i.description)
	}
	return out
}

// List of informations to display - Can be sorted
type informations []*information

func (l informations) Len() int      { return len(l) }
func (l informations) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l informations) Less(i, j int) bool {
	if l[i].stars != l[j].stars {
		return l[i].stars > l[j].stars
	}
	return l[i].name <= l[j].name
}

// Useful functions
func t(arg string) string { return gettext.Gettext(arg) }
func checkUser() {
	if e := os.Geteuid(); e == 0 {
		printError(t(MSG_NOROOT))
		os.Exit(1)
	}
}
func printError(msg interface{}) {
	fmt.Printf("\033[1;31m%v\033[m\n", msg)
}
func question(msg string, defaultValue bool) bool {
	var defstr string = "[Y/n]"
	if !defaultValue {
		defstr = "[y/N]"
	}
	fmt.Printf("\033[1;33m%s %s \033[m", msg, defstr)
	var response string
	if _, err := fmt.Scanf("%v", &response); err != nil || len(response) == 0 {
		return defaultValue
	}
	response = strings.ToLower(response)
	switch {
	case strings.HasPrefix(response, "y"):
		return true
	case strings.HasPrefix(response, "n"):
		return false
	default:
		return defaultValue
	}
}
func launchCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
func launchCommandWithResult(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	if out, err := cmd.Output(); err == nil {
		return string(out)
	}
	return ""
}
func launchRequest(debug bool, header string, searchbase string, v ...interface{}) []byte {
	search := fmt.Sprintf(searchbase, v...)
	client := &http.Client{}
	request, err := http.NewRequest("GET", search, nil)
	if err != nil {
		printError(err)
		return make([]byte, 0)
	}
	if header != "" {
		request.Header.Add("Accept", header)
	}
	response, err := client.Do(request)
	if err != nil {
		printError(err)
		return make([]byte, 0)
	}
	if debug {
		response.Write(os.Stdout)
	}
	out, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return out
}
func pathExists(path string) bool {
	_, e := os.Stat(path)
	return e == nil
}
func displayInformations(l informations, sorted bool) {
	if len(l) == 0 {
		fmt.Println(t(MSG_NOPACKAGE))
		return
	}
	if sorted {
		sort.Sort(l)
	}
	for _, i := range l {
		fmt.Println(i)
	}
}
func list(debug, checkVersions, onlyStarred bool) informations {
	out := make(informations, 0)
	ok := true
	search := SEARCH_ALL
	if onlyStarred {
		search = SEARCH_ALLST
	}
	for i := 1; ok; i++ {
		obj, e := pjson.ArrayObjectBytes(launchRequest(debug, HEADER, search, i, IDENT))
		if e != nil {
			if i == i {
				o, _ := pjson.ObjectBytes(launchRequest(debug, HEADER, search, i, IDENT))
				printError(apiError(o))
			}
			ok = false
			continue
		}
		if len(obj) == 0 {
			ok = false
			continue
		}
		for _, o := range obj {
			inf := new(information)
			if checkVersions {
				inf.update(o, NAME, DESCRIPTION, STARS)
				inf.updateLocalVersion()
			} else {
				inf.update(o, NAME, STARS)
			}
			if !onlyStarred || inf.stars > 0 {
				out = append(out, inf)
			}
		}
	}
	return out
}
func search(word string, debug, checkKcpVersion, checkLocalVersion bool) informations {
	out := make(informations, 0)
	o, e := pjson.ObjectBytes(launchRequest(debug, HEADERMATCH, SEARCH_APP, word, IDENT))
	if e != nil {
		return out
	}
	items, e := o.GetArray("items")
	if e != nil {
		printError(apiError(o))
		return out
	}
	for _, v := range items {
		if o, e := v.Object(); e == nil {
			i := new(information)
			i.update(o, NAME, DESCRIPTION, STARS)
			if checkKcpVersion {
				i.updateKcpVersion()
			}
			if checkLocalVersion {
				i.updateLocalVersion()
			}
			out = append(out, i)
		}
	}
	return out
}
func get(app string, debug bool) error {
	ok := false
	for _, i := range search(app, debug, false, false) {
		if i.name == app {
			ok = true
		}
		break
	}
	if !ok {
		return errors.New("Package not found!")
	}
	pwd, _ := os.Getwd()
	path := pwd + string(os.PathSeparator) + app
	if pathExists(path) {
		return errors.New(fmt.Sprintf(t(MSG_DIREXISTS), path))
	}
	return launchCommand("git", "clone", fmt.Sprintf(URL_REPO, app))
}
func apiError(o pjson.Object) error {
	msg, e1 := o.GetString(MESSAGE)
	doc, e2 := o.GetString(DOCUMENTATION)
	if e1 != nil || e2 != nil {
		return errors.New(t(MSG_UNKNOWN))
	}
	return errors.New(fmt.Sprintf("%s\n%s\n", msg, doc))
}

// Actions
func actionListAll(debug, fast, onlyStarred, sorted bool) {
	l := list(debug, !fast, onlyStarred)
	displayInformations(l, sorted)
}
func actionSearch(word string, debug, fast, sorted bool) {
	l := search(word, debug, !fast, true)
	displayInformations(l, sorted)
}
func actionOutOfDate(debug, sorted bool) {
	l := make(informations, 0)
	for _, app := range strings.Split(launchCommandWithResult("pacman", "-Qm"), "\n") {
		c := strings.Fields(app)
		if len(c) < 2 {
			continue
		}
		il := new(information)
		il.name = c[0]
		il.localversion = c[1]
		if !il.updateKcpVersion() || il.kcpversion == il.localversion {
			continue
		}
		description := strings.Split(launchCommandWithResult("pacman", "-Qs", il.name), "\n")
		if len(description) >= 2 {
			il.description = strings.TrimSpace(description[1])
		}
		/*
			// Unactivated code because of performance
			ok := false
			for _, ik := range search(il.name, debug, true, false) {
				if ik.name == il.name {
					if ik.kcpversion != il.localversion {
						il.kcpversion = ik.kcpversion
						il.description = ik.description
						il.stars = ik.stars
						ok = true
					}
					break
				}
			}
			if ok {
				l = append(l, il)
			}
		*/
		l = append(l, il)
	}
	displayInformations(l, sorted)
}
func actionGet(app string, debug bool) {
	if e := get(app, debug); e != nil {
		printError(e)
		os.Exit(1)
	}
}
func actionInstall(app string, debug, asdeps bool) {
	tmpDir := os.TempDir()
	os.Chdir(tmpDir)
	lck := tmpDir + string(os.PathSeparator) + KCP_LOCK
	_, e := os.Open(lck)
	if _, e := os.Open(lck); e == nil {
		printError(t(MSG_ONLYONEINSTANCE))
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
		printError(t(MSG_INTERRUPT))
		os.Exit(1)
	}()
	if e := get(app, debug); e != nil {
		printError(e)
		os.Remove(lck)
		os.Exit(1)
	}
	defer end()
	if e := os.Chdir(wDir); e != nil {
		printError(e)
		end()
		os.Exit(1)
	}
	if question(t(MSG_EDIT), true) {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		if e := launchCommand(editor, "PKGBUILD"); e != nil {
			printError(e)
			os.Exit(1)
		}
	}
	if asdeps {
		e = launchCommand("makepkg", "-si", "--asdeps")
	} else {
		e = launchCommand("makepkg", "-si")
	}
	if e != nil {
		printError(e)
		os.Exit(1)
	}
}

// Global variables
var argparser *pargs.Parser
var flag_h, flag_v, flag_l, flag_o *bool
var flag_s, flag_g, flag_i *string
var flag_fast, flag_sorted, flag_asdeps *bool
var flag_debug *bool

// Launching
func init() {
	// Init the locales
	os.Setenv("LANGUAGE", os.Getenv("LC_MESSAGES"))
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.BindTextdomain("kcp", LOCALE_DIR)
	gettext.Textdomain("kcp")

	// Init the args parser
	argparser = pargs.New(t(APP_DESCRIPTION), VERSION)
	argparser.Set(pargs.AUTHOR, AUTHOR)
	argparser.Set(pargs.SYNOPSIS, t(SYNOPSIS))
	argparser.Set(pargs.LONGDESCRIPTION, t(LONGDESCRIPTION))
	flag_h, _ = argparser.Bool("-h", "--help", t(D_HELP))
	flag_v, _ = argparser.Bool("-v", "--version", t(D_VERSION))
	flag_l, _ = argparser.Bool("-l", "--list", t(D_LIST))
	flag_o, _ = argparser.Bool("-o", "--outdated", t(D_OUTDATED))
	flag_s, _ = argparser.String("-s", "--search", t(D_SEARCH), VALUENAME, "")
	flag_g, _ = argparser.String("-g", "--get", t(D_GET), VALUENAME, "")
	flag_i, _ = argparser.String("-i", "--install", t(D_INSTALL), VALUENAME, "")
	flag_fast, _ = argparser.Bool("", "--fast", t(D_FAST))
	flag_sorted, _ = argparser.Bool("", "--sort", t(D_SORT))
	flag_asdeps, _ = argparser.Bool("", "--asdeps", t(D_ASDEPS))
	flag_debug, _ = argparser.Bool("", "--debug", "debug mode")
	argparser.GetFlag("--debug").Set(pargs.HIDDEN, true)
	argparser.Group("-h", "-v", "-l", "-o", "-s", "-g", "-i", "-l")
	argparser.Require("--fast", "-s", "-l")
	argparser.Require("--sort", "-s", "-l", "-o")
	argparser.Require("--asdeps", "-i")
}
func main() {
	checkUser()
	e := argparser.Parse(os.Args)
	switch {
	case e != nil:
		printError(e)
		fmt.Println()
		argparser.PrintHelp()
	case *flag_h:
		argparser.PrintHelp()
	case *flag_v:
		argparser.PrintVersion()
	case *flag_l:
		actionListAll(*flag_debug, *flag_fast, *flag_sorted, *flag_sorted)
	case *flag_o:
		actionOutOfDate(*flag_debug, *flag_sorted)
	case *flag_s != "":
		actionSearch(*flag_s, *flag_debug, *flag_fast, *flag_sorted)
	case *flag_g != "":
		actionGet(*flag_g, *flag_debug)
	case *flag_i != "":
		actionInstall(*flag_i, *flag_debug, *flag_asdeps)
	default:
		argparser.PrintHelp() // Should not happen
	}
}
