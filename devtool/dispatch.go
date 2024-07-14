package devtool

import (
	"fmt"
	"log"
	"net/http"
)

func Refresh() {
	_, err := http.Get(createEndpoint() + "/refresh")
	if err != nil {
		log.Fatalln(err)
	}
}

func Push(id string, props Props) error {
	req, err := http.NewRequest(http.MethodGet, createEndpoint()+"/push/"+id, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	for key, values := range props.Plugin {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	q.Add("path", props.Path)
	q.Add("name", props.Name)
	q.Add("renderTime", fmt.Sprintf("%d", props.RenderTime))
	q.Add("statusCode", fmt.Sprintf("%d", props.StatusCode))
	req.URL.RawQuery = q.Encode()
	client := new(http.Client)
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}
