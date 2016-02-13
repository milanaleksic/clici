package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	interfaceSimple   = "simple"
	interfaceAdvanced = "advanced"
)

const (
	configurationName         = "jenkins_ping.toml"
	flagForBuildingConfigFile = "make-default-config"
)

var options struct {
	Jenkins struct {
		Location string
		Jobs     []string
	}
	Application struct {
		Mock    bool
		Refresh duration
		DoLog   bool
	}
	Interface struct {
		Mode         string
		AvoidUnicode bool
	}
	CommandLine struct {
		showVersion *bool
	}
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func init() {
	options.CommandLine.showVersion = flag.Bool("version", false, "Get application version")
	buildConfFile := flag.Bool(flagForBuildingConfigFile, false, "Create default configuration file besides executable")
	flag.Parse()

	if *buildConfFile {
		asset, err := dataDefaultConfigurationToml()
		if err != nil {
			log.Fatalf("Could not create default TOML configuration file: %v", err)
		}
		if err = ioutil.WriteFile(path.Join(path.Dir(os.Args[0]), configurationName), asset.bytes, 0666); err != nil {
			log.Fatalf("Could not create default TOML configuration file: %v", err)
		}
		os.Exit(0)
	}

	if _, err := toml.DecodeFile(path.Join(path.Dir(os.Args[0]), configurationName), &options); err != nil {
		if _, err := toml.DecodeFile(configurationName, &options); err != nil {
			log.Fatalf("Failure while parsing configuration file %s: %v. If you wish to set up default configuration file, please use -%v argument",
				path.Join(path.Base(os.Args[0]), configurationName), err, flagForBuildingConfigFile)
		}
	}
}
