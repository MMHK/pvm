package main

import (
	"hjbdev/pvm/commands"
    "hjbdev/pvm/common"
    "hjbdev/pvm/theme"
	"os"
    "path/filepath"
    "runtime"
)

func main() {
	args := os.Args[1:]
	absPath, err := filepath.Abs(os.Args[0])
    basePath := filepath.ToSlash(filepath.Dir(os.Args[0]))
    if err == nil {
        basePath = filepath.ToSlash(filepath.Dir(absPath))
    }

	os := runtime.GOOS
	arch := runtime.GOARCH

	if os != "windows" {
		theme.Error("pvm currently only works on Windows.")
		return
	}

	if arch != "amd64" {
		theme.Error("pvm currently only works on 64-bit systems.")
		return
	}

	conf := common.InitConfigFile(basePath)
    confSavePath := filepath.Join(basePath, "pvm.json")
    err = conf.Save(confSavePath)
    if err != nil {
        common.Log.Error(err)
    }


	if len(args) == 0 {
		commands.Help(false)
		return
	}

	switch args[0] {
	case "help":
		commands.Help(false)
	case "list":
		commands.List(conf)
	case "path":
		commands.Path(conf)
	case "install":
		commands.Install(conf, args)
	case "use":
		commands.Use(conf, args[1:])
	default:
		commands.Help(true)
	}
}
