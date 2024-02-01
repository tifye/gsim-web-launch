package robotics

import (
	"fmt"
	"strings"
)

type Platform string

var (
	platforms []Platform
)

const (
	P25   Platform = "P25"
	P2    Platform = "P2"
	P16   Platform = "P16"
	P01G  Platform = "P01G"
	P2Z   Platform = "P2Z"
	P3    Platform = "P3"
	P005  Platform = "P005"
	P21   Platform = "P21"
	P14_2 Platform = "P14_2"
	P14_1 Platform = "P14_1"
	P005H Platform = "P005H"
	P17   Platform = "P17"
	P22   Platform = "P22"
)

func (e *Platform) String() string {
	return string(*e)
}

func GetPlatforms() []Platform {
	if platforms == nil {
		platforms = []Platform{
			P25,
			P2,
			P16,
			P01G,
			P2Z,
			P3,
			P005,
			P21,
			P14_2,
			P14_1,
			P005H,
			P17,
			P22,
		}
	}
	return platforms
}

func (e *Platform) Set(s string) error {
	s = strings.ToUpper(s)
	switch s {
	case "P25", "P2", "P16", "P01G", "P2Z", "P3", "P005", "P21", "P14_2", "P14_1", "P005H", "P17", "P22":
		*e = Platform(s)
		return nil
	default:
		return fmt.Errorf("invalid platform: %s. Must be one of %v", s, GetPlatforms())
	}
}

func (e *Platform) Type() string {
	return "Platform"
}
