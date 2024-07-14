package devtool

import (
	"fmt"
	"sort"
	"strings"
	
	"golang.org/x/exp/maps"
	
	"github.com/daarlabs/farah/ui/icon_ui"
	"github.com/daarlabs/hirokit/alpine"
	. "github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/hx"
	"github.com/daarlabs/hirokit/tempest"
)

func createTool(assetsId string, props Props) Node {
	return Html(
		Lang("en"),
		Head(
			Title(Text("Devtool")),
			Meta(CharSet("utf-8")),
			Meta(
				Name("viewport"),
				Content("width=device-width, initial-scale=1"),
			),
			Link(Rel("stylesheet"), Type("text/css"), Href(fmt.Sprintf("/.dev/.tempest/styles.%s.css", assetsId))),
			Script(Defer(), Src(fmt.Sprintf("/.dev/.tempest/scripts.%s.js", assetsId))),
		),
		Body(
			tempest.Class().BgSlate(900).TextWhite(),
			Div(
				Id(ToolId),
				createWsConnection(),
				createDevtoolContent(props),
			),
		),
	)
}

func createDevtoolContent(props Props) Node {
	var active string
	if values, ok := props.Plugin[PluginDatabase]; ok && len(values) > 0 {
		active = PluginDatabase
	}
	if values, ok := props.Plugin[PluginDebug]; ok && len(values) > 0 {
		active = PluginDebug
	}
	plugins := maps.Keys(props.Plugin)
	sort.Strings(plugins)
	if len(active) == 0 && len(plugins) > 0 {
		active = plugins[0]
	}
	return Div(
		alpine.Data(
			map[string]any{
				"active": active,
			},
		),
		Id(RequestId),
		tempest.Class().H("screen").Flex().FlexCol().Overflow("hidden").TextXs(),
		Div(
			tempest.Class().H("full").OverflowY("auto").OverflowX("hidden"),
			Range(
				plugins, func(pluginKey string, _ int) Node {
					values, ok := props.Plugin[pluginKey]
					if !ok {
						return Fragment()
					}
					plugin, ok := ToolConfig.Plugin[pluginKey]
					if !ok {
						return Fragment()
					}
					return Div(
						alpine.Cloak(),
						alpine.Show(fmt.Sprintf("active === '%s'", pluginKey)),
						Range(
							values, func(item string, _ int) Node {
								return Div(
									tempest.Class().Name(requestName).BorderB(1).BorderSlate(600).P(4).BreakAll().Whitespace("pre-wrap"),
									If(
										plugin.RowFunc == nil,
										Text(item),
									),
									If(
										plugin.RowFunc != nil,
										plugin.RowFunc(item),
									),
								)
							},
						),
					)
				},
			),
		),
		Div(
			tempest.Class().Mt("auto"),
			Div(
				tempest.Class().Flex().ItemsCenter().Gap(2).FlexWrap().FlexNone().MinH(12).Py(2).W("full").Px(2).BgSlate(700).TextXs().BorderT(1).BorderSlate(500),
				Range(
					plugins, func(pluginKey string, _ int) Node {
						values, exists := props.Plugin[pluginKey]
						if !exists || exists && len(values) == 0 {
							return Fragment()
						}
						plugin, ok := ToolConfig.Plugin[pluginKey]
						if !ok {
							return Fragment()
						}
						return Button(
							tempest.Class().Name(requestName).TextWhite().Border(1).Rounded().Px(2).H(8).Grid().PlaceItemsCenter().TextCenter(),
							Type("button"),
							alpine.Click("active = '"+pluginKey+"'"),
							alpine.Class(
								map[string]string{
									tempest.Class().Name(requestName).BgBlue(500).BorderBlue(500).String(): fmt.Sprintf(
										"active === \"%s\"", pluginKey,
									),
									tempest.Class().Name(requestName).BgTransparent().BorderSlate(500).String(): fmt.Sprintf(
										"active !== \"%s\"", pluginKey,
									),
								},
							),
							If(
								plugin.IconPath != nil,
								icon_ui.CreateIcon(
									tempest.Class().Transition().FillCurrent().Size(4),
									plugin.IconPath,
								),
							),
							If(plugin.IconPath == nil, Text(plugin.Title)),
						)
					},
				),
			),
			Div(
				tempest.Class().FlexNone().H(12).W("full").BgSlate(700).Flex().ItemsCenter().TextXs().BorderT(1).BorderSlate(500),
				Div(
					tempest.Class().FlexNone().BorderR(1).BorderSlate(500).Size(12).TextCenter().Grid().PlaceItemsCenter().FontBold().
						If(strings.HasPrefix(fmt.Sprintf("%d", props.StatusCode), "2"), tempest.Class().BgEmerald(500)).
						If(strings.HasPrefix(fmt.Sprintf("%d", props.StatusCode), "3"), tempest.Class().BgAmber(500)).
						If(strings.HasPrefix(fmt.Sprintf("%d", props.StatusCode), "4"), tempest.Class().BgRed(500)).
						If(strings.HasPrefix(fmt.Sprintf("%d", props.StatusCode), "5"), tempest.Class().BgRed(500)),
					Text(props.StatusCode),
				),
				Div(
					tempest.Class().FlexNone().BorderR(1).BorderSlate(500).H("full").Px(4).TextLeft().Grid().ItemsCenter(),
					Div(
						Div(
							tempest.Class().TextSize("9px").TextSlate(400).LhNone(),
							Text("Path:"),
						),
						Div(Text(props.Path)),
					),
				),
				Div(
					tempest.Class().FlexNone().BorderR(1).BorderSlate(500).H("full").Px(4).TextLeft().Grid().ItemsCenter(),
					Div(
						Div(
							tempest.Class().TextSize("9px").TextSlate(400).LhNone(),
							Text("Name:"),
						),
						Div(Text(props.Name)),
					),
				),
				Div(
					tempest.Class().FlexNone().Ml("auto").BorderL(1).BorderSlate(500).H("full").Px(4).TextCenter().Grid().PlaceItemsCenter().FontBold(),
					Text(formatRenderTime(props.RenderTime)),
				),
			),
		),
		Style(
			Raw(tempest.NamedStyles(requestName)),
		),
	)
}

func createWsConnection() Node {
	return Fragment(
		hx.Ext("ws"),
		CreateAttribute[string]("ws-connect")("/.dev/"),
	)
}
