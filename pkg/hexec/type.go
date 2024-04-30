package hexec

type ExecAuth interface {
	CheckAllowedExec(name string) error
	OSEnvAccept(name string) bool
}
