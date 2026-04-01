package main

import "github.com/yuhaichao/cloud189-cli/internal/commands"

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	commands.SetVersionInfo(Version, BuildTime)
	commands.Execute()
}
