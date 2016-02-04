package dbf

import (
	"log"
	"os"
)

var Log *log.Logger

func init() {
	Log = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
}
