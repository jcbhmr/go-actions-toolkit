package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var Context = newContext()

type context struct {
	Payload    any
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

func newContext() *context {
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
	return &context2{
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
}

type contextIssue struct {
	Owner  string
	Repo   string
	Number int
}

func (c *context) Issue() *contextIssue {
	var number int
	if issue, ok := c.Payload.(map[string]any)["issue"].(map[string]any); ok {
		number = issue["number"].(int)
	} else if pullRequest, ok := c.Payload.(map[string]any)["pull_request"].(map[string]any); ok {
		number = pullRequest.(map[string]any)["number"].(int)
	} else {
		number = c.Payload.(map[string]any)["number"].(int)
	}
	repoBag := c.Repo()
	owner := repoBag.Owner
	repo := repoBag.Repo
	return &contextIssue{
		Owner:  owner,
		Repo:   repo,
		Number: number,
	}
}

type contextRepo struct {
	Owner string
	Repo  string
}

func (c *context) Repo() *contextRepo {
	if repository, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok {
		parts := strings.Split(repository, "/")
		owner := parts[0]
		repo := parts[1]
		return &contextRepo{Owner: owner, Repo: repo}
	}
	if repository, ok := c.Payload["repository"].(map[string]any); ok {
		owner := repository["owner"].(map[string]any)["login"].(string)
		repo := repository["name"].(string)
		return &contextRepo{Owner: owner, Repo: repo}
	}
	panic("Context.Repo() requires a GITHUB_REPOSITORY environment variable like 'owner/repo'")
}
