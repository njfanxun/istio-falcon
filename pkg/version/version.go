package version

import (
	"fmt"
	"runtime"

	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/version"
)

var (
	gitVersion   = "v0.0.0"
	gitCommit    = "unknown"
	gitTreeState = "unknown"
	buildDate    = "unknown"
	gitMajor     = "unknown"
	gitMinor     = "unknown"
	gitTag       = "unknown"
)

type Info struct {
	GitVersion   string        `json:"gitVersion"`
	GitMajor     string        `json:"gitMajor"`
	GitMinor     string        `json:"gitMinor"`
	GitCommit    string        `json:"gitCommit"`
	GitTreeState string        `json:"gitTreeState"`
	GitTag       string        `json:"gitTag"`
	BuildDate    string        `json:"buildDate"`
	GoVersion    string        `json:"goVersion"`
	Compiler     string        `json:"compiler"`
	Platform     string        `json:"platform"`
	Kubernetes   *version.Info `json:"kubernetes,omitempty"`
}

func (info Info) String() string {
	jsonString, _ := json.Marshal(info)
	return string(jsonString)
}

func Get() Info {
	return Info{
		GitVersion:   gitVersion,
		GitMajor:     gitMajor,
		GitMinor:     gitMinor,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GitTag:       gitTag,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
