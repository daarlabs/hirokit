package hiro

import (
	"slices"
	
	"github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/socketer"
)

type Ws struct {
	clients []string
	ws      socketer.Ws
}

func (w *Ws) Client(ids ...string) *Ws {
	for _, id := range ids {
		if slices.Contains(w.clients, id) {
			continue
		}
		w.clients = append(w.clients, id)
	}
	return w
}

func (w *Ws) Broadcast(nodes ...gox.Node) error {
	w.ws.Broadcast([]byte(gox.Render(nodes...)))
	return nil
}

func (w *Ws) MustBroadcast(nodes ...gox.Node) {
	if err := w.Broadcast(nodes...); err != nil {
		panic(err)
	}
}

func (w *Ws) Send(nodes ...gox.Node) error {
	clients, err := w.ws.Find(w.clients...)
	if err != nil {
		return nil
	}
	for _, c := range clients {
		if err := c.Write([]byte(gox.Render(nodes...))); err != nil {
			return err
		}
	}
	return nil
}

func (w *Ws) MustSend(nodes ...gox.Node) {
	if err := w.Send(nodes...); err != nil {
		panic(err)
	}
}
