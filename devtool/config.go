package devtool

import (
	"fmt"
	
	"github.com/daarlabs/hirokit/tempest"
)

type Config struct {
	Port   string
	Plugin map[string]Plugin
}

var (
	ToolConfig = &Config{
		Port:   DefaultPort,
		Plugin: make(map[string]Plugin),
	}
)

func init() {
	tempest.GlobalConfig = &tempest.Config{
		FontFamily: "Sora, sans-serif",
		Font: map[string]tempest.Font{
			"sora": {
				Value: "Sora, sans-serif",
				Url:   "https://fonts.googleapis.com/css2?family=Sora:wght@100..800&display=swap",
			},
		},
		Styles:  []string{},
		Scripts: []string{},
	}
	tempest.Start()
}

func createEndpoint() string {
	return fmt.Sprintf("http://localhost:%s/.dev", ToolConfig.Port)
}
