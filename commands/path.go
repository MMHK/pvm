package commands

import (
    "fmt"
    "hjbdev/pvm/common"
    "hjbdev/pvm/theme"
    "os"
    "path/filepath"
)

func Path(conf *common.Config) {
	theme.Title("pvm: PHP Version Manager")

    if len(os.Args) > 2 {
        versionsPath, err := filepath.Abs(os.Args[2])
        if err != nil {
            versionsPath = os.Args[2]
        }
        conf.PVM_VERSIONS_PATH = filepath.ToSlash(versionsPath)
        savePath := filepath.ToSlash(filepath.Join(filepath.Dir(os.Args[0]), "pvm.json"))
        err = conf.Save(savePath)
        if err != nil {
            common.Log.Error(err)
        }

    }

	fmt.Println("PVM env:")
	fmt.Printf("PVM_PATH=%s\n", conf.PVM_PATH)
	fmt.Printf("PVM_VERISONS_PATH=%s\n", conf.PVM_VERSIONS_PATH)

	if len(os.Args) <= 2 {
	    fmt.Println("\n\nset PHP save Path")
	    fmt.Println("pvm path [versions path]")
    }
}
