package config

import "fmt"

const (
	AppName    = "polkadothub-indexer"
	AppVersion = "0.1.0"
	GitCommit  = "-"
	GoVersion  = "1.14"
)

func VersionString() string {
	return fmt.Sprintf(
		"%s %s (git: %s, %s)",
		AppName,
		AppVersion,
		GitCommit,
		GoVersion,
	)
}
