package main

import (
	banner "github.com/CrowdSurge/banner"
	cmd "github.com/adriamb/gotoma/cmd"
)

func main() {

	banner.Print("gotoma")
	cmd.ExecuteCmd()
}
