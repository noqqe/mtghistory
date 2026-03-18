package mtghistory

import (
	"os"

	"github.com/charmbracelet/log"
)

func Logger() *log.Logger {
	l := log.New(os.Stderr)
	l.SetReportTimestamp(true)
	return l
}
