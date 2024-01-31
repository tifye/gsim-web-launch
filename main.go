package main

import (
	"log"
	"os"
	"strings"

	"github.com/Tifufu/gsim-web-launch/cmd"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("could not load environment file")
	}
}

func main() {
	if strings.HasPrefix(os.Args[1], "gsim-web-launch:") {
		serial := strings.Split(os.Args[1], ":")[1]
		os.Args[1] = "-s"
		os.Args = append(os.Args, serial)
	}

	cmd.Execute(os.Args[1:])
}
