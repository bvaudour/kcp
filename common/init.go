package common

import (
	"os"
	"path"

	"github.com/bvaudour/kcp/conf"
	"github.com/leonelquinteros/gotext"
)

//Initialized at build time
var (
	Version   string
	BuildTime string

	LocaleBaseDir string
	LocaleDomain  string

	ConfigBaseDir string
	ConfigFile    string
)

//Initialized at execution time
var (
	Language      string
	DefaultEditor string
	UserBaseDir   string
	Config        *conf.Configuration
)

func setIfZero(v *string, d string) {
	if len(*v) == 0 {
		*v = d
	}
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

	// Init runtime
	Language = os.Getenv("LANGUAGE")
	DefaultEditor = os.Getenv("EDITOR")
	setIfZero(&UserBaseDir, os.Getenv("XDG_CONFIG_HOME"))

	// Default values if empty
	setIfZero(&DefaultEditor, fbDefaultEditor)
	setIfZero(&UserBaseDir, path.Join(os.Getenv("HOME"), ".config"))

	// Load the system config
	fp := path.Join(ConfigBaseDir, ConfigFile)
	var err error
	if Config, err = conf.Load(fp); err != nil {
		panic(err)
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
	gotext.Configure(LocaleBaseDir, Language, LocaleDomain)
}
