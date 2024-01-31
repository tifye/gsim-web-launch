package registry

import (
	"github.com/spf13/cobra"
)

// RegistryCmd represents the registry command
var RegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
