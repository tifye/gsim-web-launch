package clear

import (
	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/spf13/cobra"
)

func NewClearCommand(gsCli *cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newClearWinMowerFsCommand(gsCli))

	return cmd
}
