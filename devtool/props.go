package devtool

type Props struct {
	Path       string
	Name       string
	StatusCode int
	RenderTime int
	Plugin     map[string][]string
}
