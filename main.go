package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(`D:\Projects\_work\_pocs\gsim-web-launch\bin\.env`); err != nil {
		log.Println("could not load environment file")
	}
}

func main() {
	if strings.HasPrefix(os.Args[1], "gsim-web-launch:") {
		input := strings.Split(os.Args[1], ":")[1]
		parts := strings.Split(input, "/")
		serial := parts[0]
		platform := parts[1]
		args := []string{
			"-s",
			serial,
			"-p",
			platform,
		}
		os.Args = append(os.Args[:1], args...)
	}

	log.Println(os.Args)
	time.Sleep(5 * time.Second)
	cmd.Execute(os.Args[1:])
}
