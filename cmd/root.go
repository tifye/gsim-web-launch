package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/spf13/cobra"
)

var (
	serialNumber string
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(serialNumber)

		fmt.Println("Launching winmower...")
		pkg.LaunchWinMower()
		fmt.Println("Launching simulator...")
		pkg.LaunchSimulator()
		time.Sleep(3 * time.Second)
		fmt.Println("Launching test bundle...")
		pkg.RunTestBundle()
	},
}

func init() {
	cobra.MousetrapHelpText = ""
	rootCmd.AddCommand(registry.RegistryCmd)

	rootCmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the device")
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(10 * time.Second)
}
