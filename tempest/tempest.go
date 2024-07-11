package tempest

import (
	"embed"
	"fmt"
	"slices"
	"strings"
)

const (
	initialClassmapPriority = 2
)

const (
	baseCssFilename = "base.css"
)

var (
	GlobalConfig *Config
)

var (
	externalStyles string
	baseStyles     string
	scripts        string
)

var (
	defaultStyles = []string{
		"https://cdnjs.cloudflare.com/ajax/libs/modern-normalize/2.0.0/modern-normalize.min.css",
	}
	defaultScripts = []string{
		"https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js",
		"https://unpkg.com/htmx.org@1.9.12",
	}
)

//go:embed base.css
var baseCss embed.FS

func init() {
	mustReadBaseStylesFile()
	initGlobalClassmaps()
}

func Class(classes ...string) Tempest {
	return createBuilder(classes...)
}

func Start() {
	processGlobalConfig()
	injectDefaultStyles()
	injectDefaultScripts()
	processBaseStyles()
	erm := createExternalResourceManager()
	erm.mustRun()
	buildExternalStyles(erm)
	buildScripts(erm)
}

func Styles() string {
	return buildStyles()
}

func NamedStyles(name string) string {
	return buildNamedStyles(name)
}

func Scripts() string {
	return scripts
}

func injectDefaultStyles() {
	for _, s := range defaultStyles {
		if slices.Contains(GlobalConfig.Styles, s) {
			continue
		}
		GlobalConfig.Styles = append(GlobalConfig.Styles, s)
	}
}

func injectDefaultScripts() {
	for _, s := range defaultScripts {
		if slices.Contains(GlobalConfig.Scripts, s) {
			continue
		}
		GlobalConfig.Scripts = append(GlobalConfig.Scripts, s)
	}
}

func processGlobalConfig() {
	if GlobalConfig == nil {
		GlobalConfig = &Config{}
	}
	if GlobalConfig.FontSize == 0 {
		GlobalConfig.FontSize = DefaultFontSize
	}
	if GlobalConfig.Breakpoint == nil {
		GlobalConfig.Breakpoint = DefaultBreakpoints
	}
	if GlobalConfig.Container == nil {
		GlobalConfig.Container = DefaultContainer
	}
	GlobalConfig.Color = mergeConfigMap[Color](Pallete, GlobalConfig.Color)
	GlobalConfig.Shadow = mergeConfigMap[[]Shadow](BoxShadow, GlobalConfig.Shadow)
	GlobalConfig.Animation = mergeConfigMap[Animation](Animations, GlobalConfig.Animation)
	GlobalConfig.processAnimations()
	GlobalConfig.processShadows()
}

func processBaseStyles() {
	replacer := strings.NewReplacer(
		"\n", " ",
		"\t", "",
		"\r", "",
		baseFontSize, stringifyMostSuitableNumericType(GlobalConfig.FontSize)+Px,
		baseFontFamily, fmt.Sprintf("%s", GlobalConfig.FontFamily),
	)
	baseStyles = replacer.Replace(baseStyles)
}

func readBaseStylesFile() error {
	baseStylesBytes, err := baseCss.ReadFile(baseCssFilename)
	if err != nil {
		return err
	}
	baseStyles = string(baseStylesBytes)
	return nil
}

func mustReadBaseStylesFile() {
	if err := readBaseStylesFile(); err != nil {
		panic(err)
	}
}

func buildExternalStyles(erm *externalResourceManager) {
	w := new(strings.Builder)
	for _, item := range erm.styles {
		w.Write(item)
		w.WriteString("\n")
	}
	externalStyles = w.String()
}

func buildScripts(erm *externalResourceManager) {
	w := new(strings.Builder)
	for _, item := range erm.scripts {
		w.Write(item)
		w.WriteString("\n")
	}
	scripts = w.String()
}
