package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	gogithubgithub "github.com/google/go-github/v61/github"
	nethttplibrary "github.com/microsoft/kiota-http-go"
	"github.com/octokit/go-sdk/pkg/authentication"
	gosdkgithub "github.com/octokit/go-sdk/pkg/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func GetGoGithubClient(token string) (*gogithubgithub.Client, error) {
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

func GetGoSdkApiClient(token string) (*gosdkgithub.ApiClient, error) {
	provider := authentication.NewTokenProvider(
		authentication.WithAuthorizationToken(token),
		authentication.WithUserAgent("actions-toolkit.go/github"),
	)
	adapter, err := nethttplibrary.NewNetHttpRequestAdapter(provider)
	if err != nil {
		return nil, err
	}
	apiClient := gosdkgithub.NewApiClient(adapter)
	return apiClient, nil
}

func GetGithubV4Client(token string) *githubv4.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	var client *githubv4.Client
	if Context.GraphqlUrl == "https://api.github.com/graphql" {
		client = githubv4.NewClient(httpClient)
	} else {
		client = githubv4.NewEnterpriseClient(Context.GraphqlUrl, httpClient)
	}
	return client
}

var Context = mustNewContext2()

type context2 struct {
	Payload    map[string]any
	EventName  string
	Sha        string
	Ref        string
	Workflow   string
	Action     string
	Actor      string
	Job        string
	RunAttempt int
	RunNumber  int
	RunId      int
	ApiUrl     string
	ServerUrl  string
	GraphqlUrl string
}

func newContext2() (*context2, error) {
	var payload map[string]any
	if githubEventPath, ok := os.LookupEnv("GITHUB_EVENT_PATH"); ok {
		if bytes, err := os.ReadFile(githubEventPath); err == nil {
			err := json.Unmarshal(bytes, &payload)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("GITHUB_EVENT_PATH %s does not exist\n", githubEventPath)
		}
	}
	runAttempt, _ := strconv.ParseInt(os.Getenv("GITHUB_RUN_ATTEMPT"), 10, 32)
	runNumber, _ := strconv.ParseInt(os.Getenv("GITHUB_RUN_NUMBER"), 10, 32)
	runId, _ := strconv.ParseInt(os.Getenv("GITHUB_RUN_ID"), 10, 32)
	apiUrl, ok := os.LookupEnv("GITHUB_API_URL")
	if !ok {
		apiUrl = "https://api.github.com"
	}
	serverUrl, ok := os.LookupEnv("GITHUB_SERVER_URL")
	if !ok {
		serverUrl = "https://github.com"
	}
	graphqlUrl, ok := os.LookupEnv("GITHUB_GRAPHQL_URL")
	if !ok {
		graphqlUrl = "https://api.github.com/graphql"
	}
	c := &context2{
		Payload:    payload,
		EventName:  os.Getenv("GITHUB_EVENT_NAME"),
		Sha:        os.Getenv("GITHUB_SHA"),
		Ref:        os.Getenv("GITHUB_REF"),
		Workflow:   os.Getenv("GITHUB_WORKFLOW"),
		Action:     os.Getenv("GITHUB_ACTION"),
		Actor:      os.Getenv("GITHUB_ACTOR"),
		Job:        os.Getenv("GITHUB_JOB"),
		RunAttempt: int(runAttempt),
		RunNumber:  int(runNumber),
		RunId:      int(runId),
		ApiUrl:     apiUrl,
		ServerUrl:  serverUrl,
		GraphqlUrl: graphqlUrl,
	}
	return c, nil
}

func mustNewContext2() *context2 {
	c, err := newContext2()
	if err != nil {
		panic(err)
	}
	return c
}

type context2Issue struct {
	Owner  string
	Repo   string
	Number int
}

func (c *context2) Issue() *context2Issue {
	var number int
	if issue, ok := c.Payload["issue"].(map[string]any); ok {
		number = issue["number"].(int)
	} else if pullRequest, ok := c.Payload["pull_request"].(map[string]any); ok {
		number = pullRequest["number"].(int)
	} else {
		number = c.Payload["number"].(int)
	}
	repoBag := c.Repo()
	owner := repoBag.Owner
	repo := repoBag.Repo
	return &context2Issue{
		Owner:  owner,
		Repo:   repo,
		Number: number,
	}
}

type context2Repo struct {
	Owner string
	Repo  string
}

func (c *context2) Repo() *context2Repo {
	if repository, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok {
		parts := strings.Split(repository, "/")
		owner := parts[0]
		repo := parts[1]
		return &context2Repo{Owner: owner, Repo: repo}
	}
	if repository, ok := c.Payload["repository"].(map[string]any); ok {
		owner := repository["owner"].(map[string]any)["login"].(string)
		repo := repository["name"].(string)
		return &context2Repo{Owner: owner, Repo: repo}
	}
	panic("Context.Repo() requires a GITHUB_REPOSITORY environment variable like 'owner/repo'")
}
