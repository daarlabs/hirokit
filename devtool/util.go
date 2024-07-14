package devtool

import "fmt"

func formatRenderTime(renderTime int) string {
	if renderTime/1000 >= 1 {
		return fmt.Sprintf("%.2fs", float64(renderTime)/1000)
	}
	return fmt.Sprintf("%dms", renderTime)
}
