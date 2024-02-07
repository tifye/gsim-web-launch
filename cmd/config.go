package cmd

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

func setDefaults(cacheDir string) {
	viper.SetDefault("endpoints.gardenSimulatorPacket", "https://hqvrobotics.azure-api.net/gardensimulatorpacket")
	viper.SetDefault("endpoints.bundleStorage", "https://hqvrobotics.azure-api.net")

	viper.SetDefault("programs.tifConsole", filepath.Join(cacheDir, "TifApp/TifConsole.Auto.exe"))

	viper.SetDefault("directories.appCacheDir", filepath.Join(cacheDir, "gsim"))
	appCacheDir := viper.GetViper().GetString("directories.appCacheDir")
	viper.SetDefault("directories.winMowers", filepath.Join(appCacheDir, "winmower"))
	viper.SetDefault("directories.winMowerFileSystems", filepath.Join(appCacheDir, "winmower-filesystems"))
	viper.SetDefault("directories.gardenSimulatorPackets", filepath.Join(appCacheDir, "gsp"))

	viper.SetDefault("simulator.toLogNow", false)
	viper.SetDefault("simulator.screen.width", 1280)
	viper.SetDefault("simulator.screen.height", 720)
}

func initConfig() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Failed to get user cache dir", "err", err)
	}
	setDefaults(cacheDir)

	appCacheDir := viper.GetString("directories.appCacheDir")

	viper.AddConfigPath(appCacheDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create default config file
			if err := viper.WriteConfigAs(filepath.Join(appCacheDir, "config.yaml")); err != nil {
				log.Error("Failed to write default config file", "err", err)
			}
		} else {
			log.Fatal("Failed to read config file", "err", err)
		}
	}
}
