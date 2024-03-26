package cli

import (
	"flag"
	"fmt"
	"runtime/debug"
	"sync"
)

type versionCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
}

func NewVersionCmd(parent *flag.FlagSet) (*versionCmd, error) {
	nCmd := &versionCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("version", flag.ExitOnError)
	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (oc *versionCmd) Usage() {
	oc.cmd.Usage()
}

func (oc *versionCmd) Run() error {
	fmt.Println(BuildVersionString())
	return nil
}

func BuildVersionString() string {
	program := "hugoverse"

	version := "v" + CurrentVersion.String()

	bi := getBuildInfo()
	if bi == nil {
		return version
	}
	if bi.Revision != "" {
		version += "-" + bi.Revision
	}

	osArch := bi.GoOS + "/" + bi.GoArch

	date := bi.RevisionTime
	if date == "" {
		date = "unknown"
	}

	versionString := fmt.Sprintf("%s %s %s BuildDate=%s",
		program, version, osArch, date)

	return versionString
}

var (
	bInfo     *buildInfo
	bInfoInit sync.Once
)

type buildInfo struct {
	VersionControlSystem string
	Revision             string
	RevisionTime         string
	Modified             bool

	GoOS   string
	GoArch string

	*debug.BuildInfo
}

func getBuildInfo() *buildInfo {
	bInfoInit.Do(func() {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		bInfo = &buildInfo{BuildInfo: bi}

		for _, s := range bInfo.Settings {
			switch s.Key {
			case "vcs":
				bInfo.VersionControlSystem = s.Value
			case "vcs.revision":
				bInfo.Revision = s.Value
			case "vcs.time":
				bInfo.RevisionTime = s.Value
			case "vcs.modified":
				bInfo.Modified = s.Value == "true"
			case "GOOS":
				bInfo.GoOS = s.Value
			case "GOARCH":
				bInfo.GoArch = s.Value
			}
		}
	})

	return bInfo
}

var CurrentVersion = Version{
	Major:      0,
	Minor:      1,
	PatchLevel: 0,
	Suffix:     "-DEV",
}

// Version represents the Hugo build version.
type Version struct {
	Major int

	Minor int

	// Increment this for bug releases
	PatchLevel int

	// HugoVersionSuffix is the suffix used in the Hugo version string.
	// It will be blank for release versions.
	Suffix string
}

func (v Version) String() string {
	return version(v.Major, v.Minor, v.PatchLevel, v.Suffix)
}

func version(major, minor, patch int, suffix string) string {
	return fmt.Sprintf("%d.%d.%d%s", major, minor, patch, suffix)
}
