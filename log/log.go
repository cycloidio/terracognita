package log

import (
	"io"
	"io/ioutil"
	"os"
	"sync"

	kitlog "github.com/go-kit/kit/log"
	"github.com/hashicorp/terraform/helper/logging"
)

var logger kitlog.Logger
var once sync.Once

// Init initializes the log, it can only be called once,
// repetitive calls to it will not change it.
// It also set the level of vebosity of the
// Terraform logs via tflogs, if true it'll
// use the TF_LOG env variable to set it to
// Terraform
func Init(out io.Writer, tflogs bool) {
	once.Do(func() {
		w := kitlog.NewSyncWriter(out)
		logger = kitlog.NewLogfmtLogger(w)

		if !tflogs {
			os.Setenv("TF_LOG", "")
			logging.SetOutput()
		}

		logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	})
}

// Get returns the initialized logger,
// if it has not been initialized it'll
// initialize it with the default values
func Get() kitlog.Logger {
	Init(ioutil.Discard, false)
	return logger
}
