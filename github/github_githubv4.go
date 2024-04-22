//go:build toolkitgithub_githubv4

package githubv4

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	shurcoolgithubv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func GetClient(token string) *shurcoolgithubv4.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	var client *shurcoolgithubv4.Client
	if Context.GraphqlUrl == "https://api.github.com/graphql" {
		client = shurcoolgithubv4.NewClient(httpClient)
	} else {
		client = shurcoolgithubv4.NewEnterpriseClient(Context.GraphqlUrl, httpClient)
	}
	return client
}