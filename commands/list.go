package commands

import (
    "hjbdev/pvm/common"
    "hjbdev/pvm/theme"
	"log"
	"os"
	"github.com/fatih/color"
)

func List(conf *common.Config) {

	// check if .pvm/versions folder exists
	if _, err := os.Stat(conf.PVM_VERSIONS_PATH); os.IsNotExist(err) {
		theme.Error("No PHP versions installed")
		return
	}

	// get all folders in .pvm/versions
	versions, err := os.ReadDir(conf.PVM_VERSIONS_PATH)
	if err != nil {
		log.Fatalln(err)
	}

	theme.Title("Installed PHP versions")

	// print all folders
	for _, version := range versions {
		color.White("    " + version.Name())
	}
}
