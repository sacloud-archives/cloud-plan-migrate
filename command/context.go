package command

import (
	"github.com/sacloud/libsacloud/api"
)

type context struct {
	flagContext FlagContext
	client      *api.Client
	nargs       int
	args        []string
}
type Context interface {
	GetAPIClient() *api.Client
	Args() []string
	NArgs() int
	FlagContext
}

type FlagContext interface {
	IsSet(name string) bool
}

func NewContext(flagContext FlagContext, args []string, formater interface{}) Context {

	return &context{
		flagContext: flagContext,
		client:      createAPIClient(),
		args:        args,
		nargs:       len(args),
	}

}

func (c *context) GetAPIClient() *api.Client {
	return c.client
}

func (c *context) IsSet(name string) bool {
	return c.flagContext.IsSet(name)
}

func (c *context) NArgs() int {
	return c.nargs
}

func (c *context) Args() []string {
	return c.args
}
