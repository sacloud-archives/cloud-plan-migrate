package cli

import (
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
)

type FlagHandler interface {
	IsSet(name string) bool
	Set(name, value string) error
	String(name string) string
	StringSlice(name string) []string
}

func toSakuraID(id string) (int64, bool) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func hasTags(target interface{}, tags []string) bool {
	type tagHandler interface {
		HasTag(target string) bool
	}

	tagHolder, ok := target.(tagHandler)
	if !ok {
		return false
	}

	// 完全一致 + AND条件
	res := true
	for _, p := range tags {
		if !tagHolder.HasTag(p) {
			res = false
			break
		}
	}
	return res
}

func isTerminal() bool {
	is := func(fd uintptr) bool {
		return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
	}
	return is(os.Stdin.Fd()) && is(os.Stdout.Fd())
}
