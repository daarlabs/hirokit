package hiro

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"
	
	"github.com/daarlabs/hirokit/devtool"
	"github.com/daarlabs/hirokit/env"
)

type Dev interface {
	Debug(values ...any) error
	MustDebug(values ...any)
}

type devManager struct {
	query []string
	debug []string
}

func createDev() *devManager {
	return &devManager{
		query: make([]string, 0),
		debug: make([]string, 0),
	}
}

func (d *devManager) Debug(values ...any) error {
	if !env.Development() {
		return nil
	}
	for _, value := range values {
		var valueBytes []byte
		var err error
		valueBytes, err = json.Marshal(value)
		if err != nil {
			return err
		}
		d.debug = append(d.debug, string(valueBytes))
	}
	fmt.Println(strings.Join(d.debug, " "))
	return nil
}

func (d *devManager) MustDebug(values ...any) {
	if err := d.Debug(values...); err != nil {
		panic(err)
	}
}

func devtoolPush(c *ctx) {
	plugin := map[string][]string{
		devtool.PluginDatabase: c.dev.query,
		devtool.PluginDebug:    c.dev.debug,
	}
	if c.Auth().Session().MustExists() {
		session := c.Auth().Session().MustGet()
		plugin[devtool.PluginSession] = []string{
			"Id/" + fmt.Sprint(session.Id),
			"Email/" + session.Email,
			"Roles/" + strings.Join(session.Roles, ", "),
			"Ip/" + session.Ip,
			"UserAgent/" + session.UserAgent,
		}
	}
	param := c.Request().QueryMap()
	maps.Copy(param, c.Request().PathMap())
	if err := devtool.Push(
		c.State().Token(),
		devtool.Props{
			Path:       c.Request().Path(),
			Name:       c.Request().Name(),
			RenderTime: int(time.Now().Sub(c.time).Milliseconds()),
			StatusCode: c.response.StatusCode,
			Param:      param,
			Plugin:     plugin,
		},
	); err != nil {
		c.err = err
	}
}
