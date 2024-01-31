package registry

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"

	"github.com/spf13/cobra"
)

// registry/updateCmd represents the registry/update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reg key with current exe path",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		exePath := `D:\Projects\_work\_pocs\gsim-web-launch\bin\gsim-web-launch.exe`
		//exePath := `D:\Projects\_work\_pocs\gsim-web-launch\bin\run.bat`

		key, _, err := registry.CreateKey(registry.CLASSES_ROOT, `gsim-web-launch`, registry.ALL_ACCESS)
		if err != nil {
			log.Fatalf("Failed to open initial registry key: %s", err)
		}
		key.SetStringValue("", "URL: GSim Web Launch Protocol")
		key.SetStringValue("URL Protocol", "")
		key.Close()

		key, _, err = registry.CreateKey(registry.CLASSES_ROOT, `gsim-web-launch\shell\open\command`, registry.ALL_ACCESS)
		if err != nil {
			log.Fatalf("Failed to open registry key: %s", err)
		}
		key.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exePath))
		key.Close()
	},
}

func init() {
	RegistryCmd.AddCommand(updateCmd)
}
