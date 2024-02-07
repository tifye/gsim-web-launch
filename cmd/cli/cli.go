package cli

import (
	"github.com/Tifufu/gsim-web-launch/pkg/robotics"
	"github.com/spf13/viper"
)

type Cli struct {
	Config           *viper.Viper
	AppCacheDir      string
	TestingDir       string
	WinMowerRegistry *robotics.WinMowerRegistry
	BundleRegistry   *robotics.BundleRegistry
	GSPRegistry      *robotics.GSPRegistry
}
