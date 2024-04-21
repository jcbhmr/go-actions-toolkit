package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func GetIdToken(audience *string) (string, error) {
	url, ok := os.LookupEnv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if !ok {
		return "", errors.New("unable to get ACTIONS_ID_TOKEN_REQUEST_URL env variable")
	}
	if audience != nil {
		url += fmt.Sprintf("&audience=%s", *audience)
	}
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var target struct {
		Value string
	}
	err = json.NewDecoder(res.Body).Decode(&target)
	if err != nil {
		return "", err
	}
	return target.Value, nil
}
