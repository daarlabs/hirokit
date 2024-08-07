package devtool

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	
	. "github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/tempest"
)

type debugStackTrace struct {
	line    int
	path    string
	content string
}

func Recover(assets Node, err error) Node {
	stackTraceList, stackTraceErr := parseStackTrace()
	if stackTraceErr != nil {
		err = stackTraceErr
	}
	return recoverPageLayout(
		assets,
		Div(
			tempest.Class().TextRed(400).FontBold().Mb(4).TextMain(),
			Text("Error: "+err.Error()),
		),
		Div(
			tempest.Class().Grid().Gap(2).Mb(6),
			Range(
				stackTraceList, func(item debugStackTrace, _ int) Node {
					return Div(
						tempest.Class().Border(1).BorderSlate(500).Rounded().P(6).Lh(6),
						Div(
							tempest.Class().TextSky(300).FontSemibold().Mb(2),
							Text(item.path+":"+strconv.Itoa(item.line)),
						),
						Div(
							tempest.Class().TextAmber(200).BgAmber(900).Py(2).Px(4),
							Text(item.content),
						),
					)
				},
			),
		),
		Div(
			tempest.Class().TextCenter(),
			A(
				tempest.Class().TextWhite().Underline().NoUnderline(tempest.Hover()),
				Href(""),
				Text("Refresh"),
			),
		),
	)
}

func parseStackTrace() ([]debugStackTrace, error) {
	var buf [4096]byte
	result := make([]debugStackTrace, 0)
	n := runtime.Stack(buf[:], false)
	stackTrace := string(buf[:n])
	wd, err := os.Getwd()
	if err != nil {
		return result, err
	}
	for _, item := range strings.Split(stackTrace, "\n") {
		if strings.Contains(item, "hirokit") ||
			strings.Contains(item, "/go/") ||
			strings.Contains(item, "panic") ||
			strings.Contains(item, "net/http") ||
			strings.Contains(item, "goroutine") {
			continue
		}
		item = strings.TrimSpace(item)
		if len(item) == 0 || !strings.Contains(item, ".go:") {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(item, wd), " ")
		if len(parts) < 1 {
			continue
		}
		pathLine := parts[0]
		if !strings.Contains(pathLine, ":") {
			continue
		}
		path := pathLine[:strings.LastIndex(pathLine, ":")]
		line, err := strconv.Atoi(pathLine[strings.LastIndex(pathLine, ":")+1:])
		if err != nil {
			return result, err
		}
		fileBytes, err := os.ReadFile(wd + path)
		if err != nil {
			return result, err
		}
		for i, fileLineContent := range strings.Split(string(fileBytes), "\n") {
			if i != line-1 {
				continue
			}
			result = append(
				result, debugStackTrace{
					line:    line,
					path:    path,
					content: strings.TrimSpace(fileLineContent),
				},
			)
		}
	}
	return result, nil
}

func recoverPageLayout(assets Node, nodes ...Node) Node {
	return Html(
		Head(
			Title(Text("Recovered error")),
			Meta(
				Name("viewport"),
				Content("width=device-width, initial-scale=1"),
			),
			Raw(
				`
				<link rel="apple-touch-icon" sizes="180x180" href="/public/favicon/apple-touch-icon.png">
				<link rel="icon" type="image/png" sizes="32x32" href="/public/favicon/favicon-32x32.png">
				<link rel="icon" type="image/png" sizes="16x16" href="/public/favicon/favicon-16x16.png">
				<link rel="manifest" href="/public/favicon/site.webmanifest">
				<link rel="mask-icon" href="/public/favicon/safari-pinned-tab.svg" color="#5bbad5">
				<link rel="shortcut icon" href="/public/favicon/favicon.ico">
				<meta name="msapplication-TileColor" content="#00aba9">
				<meta name="msapplication-config" content="/public/favicon/browserconfig.xml">
				<meta name="theme-color" content="#ffffff">
			`,
			),
			assets,
		),
		Body(
			tempest.Class().BgSlate(900).TextWhite().TextXs().Grid().PlaceItemsCenter().
				H("screen").W("screen").Overflow("auto"),
			Div(
				Fragment(nodes...),
			),
		),
	)
}
