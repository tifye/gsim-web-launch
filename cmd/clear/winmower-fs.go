package clear

import (
	"os"
	"path/filepath"

	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func newClearWinMowerFsCommand(gsCli *cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "winmower-fs",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := os.RemoveAll(filepath.Join(gsCli.AppCacheDir, "winmower-filesystems"))
			if err != nil {
				log.Error("Failed to remove winmower filesystems", "error", err)
			}
		},
	}
	return cmd
}
