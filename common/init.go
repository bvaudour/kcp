package common

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"git.kaosx.ovh/benjamin/ini"
	"github.com/leonelquinteros/gotext"
)

//go:embed resources/kcp.conf
var embedConf []byte

// Initialized at build time
var (
	Version   string
	BuildTime string

	LocaleBaseDir string
	LocaleDomain  string

	ConfigBaseDir string
	ConfigFile    string

	GitDomain    string
	Organization string
	User         string
	Password     string
	Token        string
)

// Initialized at execution time
var (
	Language      string
	DefaultEditor string
	UserBaseDir   string
	Config        *ini.IniFile
	Exceptions    []string

	availableLocales []string
)

func setIfZero(v *string, d string) {
	if len(*v) == 0 {
		*v = d
	}
}

func getAvailableLocales() {
	dirs, err := os.ReadDir(LocaleBaseDir)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		locale := dir.Name()
		file := path.Join(LocaleBaseDir, locale, "LC_MESSAGES", LocaleDomain+".mo")
		if FileExists(file) {
			availableLocales = append(availableLocales, locale)
		}
	}

	return
}

func checkLanguage() bool {
	// If English, it' valid.
	if Language == "en" {
		return true
	}

	// No modification
	if slices.Contains(availableLocales, Language) {
		return true
	}
	// Trying to replace fr_FR.UTF-8 by fr_FR, for example
	if i := strings.Index(Language, "."); i > 0 {
		Language = Language[:i]
		if slices.Contains(availableLocales, Language) {
			return true
		}
	}
	// Trying base: fr_FR by fr, for example
	if i := strings.Index(Language, "_"); i > 0 {
		if slices.Contains(availableLocales, Language) {
			return true
		}
	}
	// Trying complete base: fr by fr_FR
	cl := Language
	Language = fmt.Sprintf("%s_%s", Language, strings.ToUpper(Language))
	if slices.Contains(availableLocales, Language) {
		return true
	}

	// Trying complete fr by fr_*
	Language = cl
	prefix := Language + "_"
	for _, locale := range availableLocales {
		if strings.HasPrefix(locale, prefix) {
			Language = locale
			return true
		}
	}
	return false
}

func initLanguage() {
	getAvailableLocales()

	if Language = Config.Get("main.language"); Language != "" && checkLanguage() {
		return
	}
	if Language = os.Getenv("LANGUAGE"); Language != "" && checkLanguage() {
		return
	}
	if Language = os.Getenv("LANG"); Language != "" && checkLanguage() {
		return
	}
	Language = "en"
}

func init() {
	// Init buildInfo
	setIfZero(&Version, fbVersion)
	setIfZero(&BuildTime, TimestampToString(Now()))

	// Init gettext
	setIfZero(&LocaleBaseDir, fbLocaleBaseDir)
	setIfZero(&LocaleDomain, fbLocaleDomain)

	// Init config
	setIfZero(&ConfigBaseDir, fbConfigBaseDir)
	setIfZero(&ConfigFile, fbConfigFile)
	setIfZero(&UserBaseDir, fbConfigUserDir)
	setIfZero(&Organization, fbOrganization)

	// Init runtime
	DefaultEditor = os.Getenv("EDITOR")
	setIfZero(&UserBaseDir, os.Getenv("XDG_CONFIG_HOME"))

	// Default values if empty
	setIfZero(&DefaultEditor, fbDefaultEditor)
	setIfZero(&UserBaseDir, path.Join(os.Getenv("HOME"), ".config"))

	// Load the system config
	fp := path.Join(ConfigBaseDir, ConfigFile)
	Config = ini.NewFile(string(embedConf), fp)
	Config.Load()

	// Load the user config
	userDir := Config.Get("main.configDir")
	setIfZero(&userDir, fbLocaleDomain)
	UserBaseDir = JoinIfRelative(UserBaseDir, userDir)
	if !FileExists(UserBaseDir) {
		os.MkdirAll(UserBaseDir, 0755)
	}
	userfp := path.Join(UserBaseDir, ConfigFile)
	Config.Path = userfp
	Config.Load()
	Config.Save()

	// Load locales
	initLanguage()
	gotext.Configure(LocaleBaseDir, Language, LocaleDomain)

	// Load custom git config
	domain, org := Config.Get("git.domain"), Config.Get("git.organization")
	user, passwd, token := Config.Get("git.user"), Config.Get("git.password"), Config.Get("git.token")
	if domain != "" {
		GitDomain = domain
	}
	if org != "" {
		Organization = org
	}
	if token != "" {
		Token = token
	} else if user != "" && passwd != "" {
		User, Password = user, passwd
	}

	// Load exceptions
	exceptionsFile := Config.Get("pckcp.exceptionsFile")
	if exceptionsFile != "" {
		p := path.Join(userDir, exceptionsFile)
		if f, err := os.ReadFile(p); err == nil {
			sc := bufio.NewScanner(bytes.NewReader(f))
			for sc.Scan() {
				Exceptions = append(Exceptions, sc.Text())
			}
		}
		p = path.Join(ConfigBaseDir, exceptionsFile)
		if f, err := os.ReadFile(p); err == nil {
			sc := bufio.NewScanner(bytes.NewReader(f))
			for sc.Scan() {
				Exceptions = append(Exceptions, sc.Text())
			}
		}
	}
}
