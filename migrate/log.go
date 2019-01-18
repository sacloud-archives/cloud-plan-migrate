package migrate

type Logger interface {
	Printf(string, ...interface{})
}
