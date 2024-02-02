package cli

import "github.com/Tifufu/gsim-web-launch/pkg/robotics"

type Cli struct {
	AppCacheDir      string
	TestingDir       string
	WinMowerRegistry *robotics.WinMowerRegistry
}
