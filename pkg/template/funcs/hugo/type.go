package hugo

type Info interface {
	Version
	Language
	Host
	Fs
}

type Version interface {
	Version() string
}

type Language interface {
	IsMultilingual() bool
}

type Host interface {
	IsMultihost() bool
}

type Fs interface {
	WorkingDir() string
}
