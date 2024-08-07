package socketer

import (
	"fmt"
	"net/http"
	"slices"
	
	"github.com/gorilla/websocket"
)

type Ws interface {
	Broadcast(bytes []byte)
	Find(id ...string) ([]Client, error)
	FindOne(id string) (Client, error)
	Upgrade(req *http.Request, res http.ResponseWriter, id string) error
	OnRead(fn func(bytes []byte)) Ws
	
	MustFind(id ...string) []Client
	MustFindOne(id string) Client
	MustUpgrade(req *http.Request, res http.ResponseWriter, id string)
}

type ws struct {
	config   Config
	hub      *hub
	upgrader *websocket.Upgrader
	onRead   func(bytes []byte)
}

func New(config ...Config) Ws {
	cfg := defaultConfig
	if len(config) > 0 {
		if config[0].ReadBufferSize > 0 {
			cfg.ReadBufferSize = config[0].ReadBufferSize
		}
		if config[0].WriteBufferSize > 0 {
			cfg.WriteBufferSize = config[0].WriteBufferSize
		}
		if config[0].ReadLimit > 0 {
			cfg.ReadLimit = config[0].ReadLimit
		}
		if config[0].WriteLimit > 0 {
			cfg.WriteLimit = config[0].WriteLimit
		}
	}
	h := createHub()
	go h.run()
	return &ws{
		config:   cfg,
		hub:      h,
		upgrader: createUpgrader(cfg),
	}
}

func (s *ws) Broadcast(bytes []byte) {
	s.hub.broadcast <- bytes
}

func (s *ws) Find(id ...string) ([]Client, error) {
	result := make([]Client, 0)
	for c := range s.hub.clients {
		if slices.Contains(id, c.id) {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return result, ErrorInvalidClient
	}
	return result, nil
}

func (s *ws) MustFind(id ...string) []Client {
	c, err := s.Find(id...)
	if err != nil {
		panic(err)
	}
	return c
}

func (s *ws) FindOne(id string) (Client, error) {
	for c := range s.hub.clients {
		if c.id == id {
			return c, nil
		}
	}
	return nil, ErrorInvalidClient
}

func (s *ws) MustFindOne(id string) Client {
	c, err := s.FindOne(id)
	if err != nil {
		panic(err)
	}
	return c
}

func (s *ws) OnRead(fn func(bytes []byte)) Ws {
	s.onRead = fn
	return s
}

func (s *ws) Upgrade(req *http.Request, res http.ResponseWriter, id string) error {
	conn, err := s.upgrader.Upgrade(res, req, nil)
	if err != nil {
		return err
	}
	defer func() {
		if e := recover(); e != nil {
			if err := conn.Close(); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(res, fmt.Sprintf("%v", e), http.StatusInternalServerError)
		}
	}()
	c := createClient(s.config, s.hub, conn, id)
	s.hub.register <- c
	go func() {
		for {
			select {
			case message := <-s.hub.read:
				if s.onRead != nil {
					s.onRead(message)
				}
			}
		}
		
	}()
	go c.watchRead()
	go c.watchWrite()
	return nil
}

func (s *ws) MustUpgrade(req *http.Request, res http.ResponseWriter, id string) {
	err := s.Upgrade(req, res, id)
	if err != nil {
		panic(err)
	}
}
