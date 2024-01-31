package main

import (
	"os"
	"strings"

	"github.com/Tifufu/gsim-web-launch/cmd"
)

func main() {
	if strings.HasPrefix(os.Args[1], "gsim-web-launch:") {
		serial := strings.Split(os.Args[1], ":")[1]
		os.Args[1] = "-s"
		os.Args = append(os.Args, serial)
	}

	cmd.Execute(os.Args[1:])
}
