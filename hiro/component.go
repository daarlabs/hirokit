package hiro

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
	
	"github.com/iancoleman/strcase"
	
	"github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/hx"
)

type MandatoryComponent interface {
	gox.Node
	Name() string
	Mount()
}

type Component struct {
	Ctx `json:"-"`
}

type Interaction struct {
	c       Component
	trigger string
	target  string
	action  string
	swap    string
	values  string
}

type component struct {
	ct     MandatoryComponent
	ctx    *ctx
	v      reflect.Value
	t      reflect.Type
	route  *Route
	action string
}

type componentCtx struct {
	name string
}

var (
	componentType       = reflect.TypeOf(Component{})
	componentExpiration = 7 * 24 * time.Hour
)

func createComponent(ct MandatoryComponent, ctx *ctx, route *Route, action string) *component {
	c := &component{
		ct:     ct,
		ctx:    ctx,
		t:      reflect.TypeOf(ct),
		v:      reflect.ValueOf(ct),
		route:  route,
		action: action,
	}
	return c
}

func (c Component) On(trigger string) *Interaction {
	return &Interaction{
		c:       c,
		trigger: trigger,
		swap:    hx.SwapNone,
	}
}

func (i *Interaction) Node() gox.Node {
	return gox.Fragment(
		gox.If(len(i.action) > 0, hx.Get(i.action)),
		gox.If(len(i.target) > 0, hx.Target(i.target)),
		gox.If(len(i.trigger) > 0, hx.Trigger(i.trigger)),
		gox.If(len(i.swap) > 0, hx.Swap(i.swap)),
		gox.If(len(i.values) > 0, hx.Vals(i.values)),
	)
}

func (i *Interaction) Action(action string, args ...Map) *Interaction {
	i.action = i.c.Generate().Action(action, args...)
	return i
}

func (i *Interaction) Replace(target string) *Interaction {
	i.createTarget(target)
	i.swap = hx.SwapOuterHtml
	return i
}

func (i *Interaction) Prepend(target string) *Interaction {
	i.createTarget(target)
	i.swap = hx.SwapBeforeBegin
	return i
}

func (i *Interaction) Append(target string) *Interaction {
	i.createTarget(target)
	i.swap = hx.SwapBeforeEnd
	return i
}

func (i *Interaction) Delete(target string) *Interaction {
	i.createTarget(target)
	i.swap = hx.SwapDelete
	return i
}

func (i *Interaction) With(values Map) *Interaction {
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return i
	}
	i.values = string(valuesBytes)
	return i
}

func (i *Interaction) createTarget(target string) {
	i.target = hx.HashId(strings.TrimPrefix(target, "#"))
}

func (c *component) render() gox.Node {
	var methodName string
	var shouldCallAction bool
	isActionRequest := len(c.action) > 0
	if isActionRequest {
		methodName, shouldCallAction = c.getAction()
	}
	c.mustGet()
	c.injectContext()
	c.ct.Mount()
	if shouldCallAction {
		c.callAction(methodName)
		return gox.Fragment()
	}
	return c.ct.Node()
}

func (c *component) getAction() (string, bool) {
	action := c.action
	parts := strings.Split(action, namePrefixDivider)
	if len(parts) < 3 {
		return "", false
	}
	n := len(parts)
	compName := parts[n-2]
	methodName := parts[n-1]
	return methodName, compName == strcase.ToKebab(c.ct.Name()) && c.v.MethodByName(methodName).IsValid()
}

func (c *component) callAction(methodName string) {
	method := c.v.MethodByName(methodName)
	if !method.IsValid() {
		return
	}
	methodResult := method.Call([]reflect.Value{})
	c.save()
	if len(methodResult) == 0 {
		return
	}
	*c.ctx.write = false
	switch r := methodResult[0].Interface().(type) {
	case error:
		c.ctx.err = r
	}
}

func (c *component) injectContext() {
	compField := c.v.Elem().FieldByName(componentType.Name())
	if !compField.IsValid() {
		return
	}
	compCtx := *c.ctx
	compCtx.component = &componentCtx{
		name: strcase.ToKebab(c.ct.Name()),
	}
	compField.Set(reflect.ValueOf(Component{Ctx: &compCtx}))
}

func (c *component) get() error {
	ct, ok := c.ctx.state.Components[c.getFullname()]
	if !ok {
		return nil
	}
	bytes, err := json.Marshal(ct)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &c.ct)
}

func (c *component) mustGet() {
	err := c.get()
	if err != nil {
		panic(err)
	}
}

func (c *component) save() {
	name := c.getFullname()
	c.ctx.state.Components[name] = c.ct
	c.ctx.state.ComponentsExpiration[name] = time.Now().Add(componentExpiration)
	c.ctx.state.mustSave()
}

func (c *component) getFullname() string {
	return createDividedName(c.route.Name, strcase.ToKebab(c.ct.Name()))
}
