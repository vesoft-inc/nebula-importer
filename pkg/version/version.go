package version

import (
	"fmt"
	"runtime"
)

const undefined = "<undefined>"

var (
	buildVersion    = undefined
	buildCommit     = undefined
	buildCommitDate = undefined
	buildDate       = undefined
)

type Version struct {
	Version    string
	Commit     string
	CommitDate string
	BuildDate  string
	GoVersion  string
	Platform   string
}

func GetVersion() *Version {
	return &Version{
		Version:    buildVersion,
		Commit:     buildCommit,
		CommitDate: buildCommitDate,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (v *Version) String() string {
	return fmt.Sprintf(`Version:    %s
Commit:     %s
CommitDate: %s
BuildDate:  %s
GoVersion:  %s
Platform:   %s
`,
		v.Version,
		v.Commit,
		v.CommitDate,
		v.BuildDate,
		v.GoVersion,
		v.Platform,
	)
}
