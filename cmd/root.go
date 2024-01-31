package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Launching winmower...")
		pkg.LaunchWinMower()
		fmt.Println("Launching simulator...")
		pkg.LaunchSimulator()
		time.Sleep(3 * time.Second)
		fmt.Println("Launching test bundle...")
		pkg.RunTestBundle()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
