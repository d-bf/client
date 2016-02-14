package dbf

import (
	"io"
	"log"
	"os"
)

var Log *log.Logger

func InitLog() {
	logFile, err := os.OpenFile(PathData+"error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		Log = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
		Log.Printf("%s", err)
	} else {
		Log = log.New(io.MultiWriter(os.Stderr, logFile), "", log.LstdFlags|log.Lshortfile)
	}
}
