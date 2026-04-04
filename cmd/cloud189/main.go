package main

import "github.com/yuhaichao/cloud189-cli/internal/commands"

var (
	Version   = "v1.6.0"
	BuildTime = "unknown"
)

func main() {
	commands.SetVersionInfo(Version, BuildTime)
	commands.Execute()
}
