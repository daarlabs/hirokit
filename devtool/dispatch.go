package devtool

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
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
	qp := make([]string, 0)
	for key, value := range props.Param {
		if !slices.Contains(qp, key) {
			q.Add(PluginParam, key)
			qp = append(qp, key)
		}
		switch v := value.(type) {
		case []string:
			for _, item := range v {
				valueBytes, err := json.Marshal(item)
				if err != nil {
					continue
				}
				q.Set(key, string(valueBytes))
			}
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
