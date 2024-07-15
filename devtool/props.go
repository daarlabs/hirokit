package devtool

type Props struct {
	Path       string
	Name       string
	StatusCode int
	RenderTime int
	Param      map[string]any
	Plugin     map[string][]string
}
