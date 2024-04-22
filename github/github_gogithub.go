//go:build !(toolkitgithub_githubv4 || toolkitgithub_gosdk)

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	gogithubgithub "github.com/google/go-github/v61/github"
)

func GetClient(token string) (*gogithubgithub.Client, error) {
	client := gogithubgithub.NewClient(nil)
	if Context.ApiUrl != "https://api.github.com" {
		var err error
		client, err = client.WithEnterpriseURLs(Context.ApiUrl, Context.ApiUrl)
		if err != nil {
			return nil, err
		}
	}
	client = client.WithAuthToken(token)
	return client, nil
}