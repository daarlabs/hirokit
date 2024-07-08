package translator

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Translator interface {
	Translate(langCode, key string, args ...map[string]any) string
	Extend(langCode string, locales map[string]string) Translator
}

type translator struct {
	translates map[string]map[string]string
}

const (
	Json = "json"
	Yaml = "yaml"
	Toml = "toml"
)

func New(config ...Config) Translator {
	t := &translator{
		translates: make(map[string]map[string]string),
	}
	for _, c := range config {
		if len(c.Dir) == 0 {
			return t
		}
		if _, err := os.Stat(c.Dir); os.IsNotExist(err) {
			panic(ErrorInvalidDir)
		}
		err := t.walk(c)
		if err != nil {
			panic(err)
		}
	}
	return t
}

func (t *translator) Extend(langCode string, locales map[string]string) Translator {
	langTranslates, ok := t.translates[langCode]
	if !ok {
		t.translates[langCode] = make(map[string]string)
	}
	for k, v := range locales {
		langTranslates[k] = v
	}
	return t
}

func (t *translator) Translate(langCode, key string, args ...map[string]any) string {
	langTranslates, ok := t.translates[langCode]
	if !ok {
		return key
	}
	translate, ok := langTranslates[key]
	if !ok {
		return key
	}
	translate = replaceArgs(translate, args...)
	return translate
}

func (t *translator) walk(c Config) error {
	if err := filepath.Walk(
		c.Dir, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(info.Name(), "."+c.FileType) {
				return nil
			}
			lang := strings.TrimSuffix(info.Name(), "."+c.FileType)
			if t.translates[lang] == nil {
				t.translates[lang] = make(map[string]string)
			}
			dir := strings.TrimPrefix(c.Dir, "./")
			subpath := strings.TrimPrefix(strings.TrimSuffix(path, info.Name()), dir)
			subpath = strings.TrimPrefix(subpath, "/")
			subpath = strings.TrimSuffix(subpath, "/")
			if err := t.read(c, lang, path, createKeyPrefixFromPath(subpath)); err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}

func (t *translator) read(c Config, lang, path, prefix string) error {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fileData := make(map[string]any)
	switch c.FileType {
	case Json:
		if err := json.Unmarshal(fileBytes, &fileData); err != nil {
			return err
		}
	case Toml:
		if err := toml.Unmarshal(fileBytes, &fileData); err != nil {
			return err
		}
	case Yaml:
		if err := yaml.Unmarshal(fileBytes, &fileData); err != nil {
			return err
		}
	}
	return t.parse(lang, prefix, fileData)
}

func (t *translator) parse(lang, prefix string, data map[string]any) error {
	hasPrefix := len(prefix) > 0
	for dataKey, item := range data {
		var key string
		if hasPrefix {
			key = fmt.Sprintf("%s.%v", prefix, dataKey)
		}
		if !hasPrefix {
			key = fmt.Sprintf("%v", dataKey)
		}
		switch d := item.(type) {
		case string:
			t.translates[lang][key] = fmt.Sprintf("%v", item)
		case map[string]any:
			if err := t.parse(lang, key, d); err != nil {
				return err
			}
		}
	}
	return nil
}
