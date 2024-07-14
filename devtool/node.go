package devtool

import (
	. "github.com/daarlabs/hirokit/gox"
)

func CreateNode() Node {
	return Div(
		Id(ConnectorId),
		createWsConnection(),
		Div(
			Id(RefreshId),
		),
		A(
			Style(Raw(`display: grid; place-items: center; position:fixed; bottom:4px; right:4px; font-size: 10px; color: #ffffff; z-index: 9999; padding:8px; background-color: rgba(0,0,0,.7);`)),
			Href("/.dev/tool/"),
			Target("_blank"),
			Svg(
				Style(Raw(`width: 16px; height: 16px;`)),
				Xmlns("http://www.w3.org/2000/svg"),
				ViewBox("0 0 24 24"),
				Fill("white"),
				Path(
					D(`M17 8V2H20C20.5523 2 21 2.44772 21 3V7C21 7.55228 20.5523 8 20 8H17ZM15 22C15 22.5523 14.5523 23 14 23H10C9.44772 23 9 22.5523 9 22V8H2.5V6.07437C2.5 5.7187 2.68891 5.3898 2.99613 5.21059L8.5 2H15V22Z`),
				),
			),
		),
	)
}
