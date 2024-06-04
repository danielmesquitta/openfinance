package container

import (
	"log"
	"os"
	"path/filepath"
)

var CWD string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	CWD = filepath.Join(wd, "..", "..", "..")
}
