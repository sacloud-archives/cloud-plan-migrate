package command

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

type Option struct {
	AccessToken       string
	AccessTokenSecret string
	Zone              string
	Timeout           int
	AcceptLanguage    string
	RetryMax          int
	RetryIntervalSec  int64
	Zones             []string
	APIRootURL        string
	TraceMode         bool
	Format            string
	DefaultOutputType string
	NoColor           bool
	In                *os.File
	Out               io.Writer
	Progress          io.Writer
	Err               io.Writer
	Validated         bool
	Valid             bool
	ValidationResults []error
}

var GlobalOption = &Option{
	In:       os.Stdin,
	Out:      colorable.NewColorableStdout(),
	Progress: colorable.NewColorableStderr(),
	Err:      colorable.NewColorableStderr(),
}

var (
	DefaultZone       = "is1a"
	DefaultOutputType = "table"
)

func init() {
	if !(isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) {
		GlobalOption.Progress = ioutil.Discard
	}
}

func (o *Option) Validate(skipAuth bool) []error {
	var errs []error

	// token/secret
	needAuth := !skipAuth
	if needAuth {
		errs = append(errs, ValidateRequired("token", o.AccessToken)...)
		errs = append(errs, ValidateRequired("secret", o.AccessTokenSecret)...)
		errs = append(errs, ValidateRequired("zone", o.Zone)...)
		errs = append(errs, ValidateInStrValues("zone", o.Zone, "is1a")...)
	}

	o.Validated = true
	o.Valid = len(errs) == 0
	o.ValidationResults = errs

	return errs
}
