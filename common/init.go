package common

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bvaudour/kcp/conf"
	"github.com/leonelquinteros/gotext"
)

//go:embed kcp.conf
var embedConf []byte

// Initialized at build time
var (
	Version   string
	BuildTime string

	LocaleBaseDir string
	LocaleDomain  string

	ConfigBaseDir string
	ConfigFile    string

	Organization string
	User         string
	Password     string
)

// Initialized at execution time
var (
	Language      string
	DefaultEditor string
	UserBaseDir   string
	Config        *conf.Configuration
	Exceptions    = make(map[string]bool)
)

func setIfZero(v *string, d string) {
	if len(*v) == 0 {
		*v = d
	}
}

func initLanguage() {
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

func languageFileExist() bool {
	p := path.Join(LocaleBaseDir, Language, "LC_MESSAGES", LocaleDomain+".mo")
	return FileExists(p)
}

func checkLanguage() bool {
	// No modification
	if languageFileExist() {
		return true
	}
	// Trying to replace fr_FR.UTF-8 by fr_FR, for example
	if i := strings.Index(Language, "."); i > 0 {
		Language = Language[:i]
		if languageFileExist() {
			return true
		}
	}
	// Trying base: fr_FR by fr, for example
	if i := strings.Index(Language, "_"); i > 0 {
		if languageFileExist() {
			return true
		}
	}
	// Trying complete base: fr by fr_FR
	cl := Language
	Language = fmt.Sprintf("%s_%s", Language, strings.ToUpper(Language))
	if languageFileExist() {
		return true
	}
	Language = cl
	//@TODO: Trying to search in similar locale (ie. if there is one fr_*)
	return false
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

	// Load the default config
	Config = conf.Parse(bytes.NewReader(embedConf))

	// Load the system config
	fp := path.Join(ConfigBaseDir, ConfigFile)
	if cs, err := conf.Load(fp); err == nil {
		Config.Fusion(cs)
	}

	// Load the user config
	userDir := Config.Get("main.configDir")
	setIfZero(&userDir, fbLocaleDomain)
	UserBaseDir = JoinIfRelative(UserBaseDir, userDir)
	if !FileExists(UserBaseDir) {
		os.MkdirAll(UserBaseDir, 0755)
	}
	userfp := path.Join(UserBaseDir, ConfigFile)
	if cu, err := conf.Load(userfp); err == nil {
		Config.Fusion(cu)
	}
	conf.Save(userfp, Config)

	// Load locales
	initLanguage()
	gotext.Configure(LocaleBaseDir, Language, LocaleDomain)

	// Load custom github config
	org, user, passwd := Config.Get("github.organization"), Config.Get("github.user"), Config.Get("github.password")
	if org != "" {
		Organization = org
	}
	if user != "" && passwd != "" {
		User, Password = user, passwd
	}

	// Load exceptions
	exceptionsFile := Config.Get("pckcp.exceptionsFile")
	if exceptionsFile != "" {
		p := path.Join(userDir, exceptionsFile)
		if f, err := os.ReadFile(p); err == nil {
			sc := bufio.NewScanner(bytes.NewReader(f))
			for sc.Scan() {
				Exceptions[sc.Text()] = true
			}
		}
		p = path.Join(ConfigBaseDir, exceptionsFile)
		if f, err := os.ReadFile(p); err == nil {
			sc := bufio.NewScanner(bytes.NewReader(f))
			for sc.Scan() {
				Exceptions[sc.Text()] = true
			}
		}
	}
}
