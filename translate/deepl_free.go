package translate

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
)

func DeeplFreeTranslate(q, source, target string) (result string, err error) {
	var resp *http.Response
	if resp, err = fetch.Post(
		"http://ubuntu:18888/translate/deepl",
		fetch.WithJSONBody(map[string]string{
			"content": q,
		}),
		fetch.WithHeader("Content-Type", "application/json"),
	); err != nil {
		return
	}
	defer resp.Body.Close()
	data := struct {
		Content string `json:"content"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}
	fmt.Println(data.Content)
	return data.Content, nil
}
