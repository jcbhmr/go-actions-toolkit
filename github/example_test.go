package github_test

import (
	"context"
	"fmt"
	"os"

	"github.com/jcbhmr/actions-toolkit.go/github"
)

func unwrap1[A any](a A, err error) A {
	if err != nil {
		panic(err)
	}
	return a
}

func unwrap2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

func init() {
	if _, ok := os.LookupEnv("GITHUB_TOKEN"); !ok {
		panic("no GITHUB_TOKEN")
	}
	if _, ok := os.LookupEnv("GITHUB_REPOSITORY"); !ok {
		panic("no GITHUB_REPOSITORY")
	}
}

func Example() {
	token := os.Getenv("GITHUB_TOKEN")
	client := unwrap1(github.GetGoGithub(token))
	repo := github.Context.Repo()
	info, _ := unwrap2(client.Repositories.Get(context.Background(), repo.Owner, repo.Repo))
	fmt.Printf("info.GetFullName()=%s\n", info.GetFullName())
	fmt.Printf("info.GetDescription()=%s\n", info.GetDescription())
	// Output:
	// info.GetFullName()=jcbhmr/actions-toolkit.go
	// info.GetDescription()=🐿️ GitHub Actions toolkit for your Go-based GitHub Actions
}

func ExampleGetGoGithub() {
	token := os.Getenv("GITHUB_TOKEN")
	client := unwrap1(github.GetGoGithub(token))
	query := "is:issue repo:nodejs/node is:open Proposal for single-mode packages with optional fallbacks for older versions of node"
	issues, _ := unwrap2(client.Search.Issues(context.Background(), query, nil))
	fmt.Printf("issues.GetTotal()=%d\n", issues.GetTotal())
	fmt.Printf("issues.Issues[0].GetNumber()=%d\n", issues.Issues[0].GetNumber())
	// Output:
	// issues.GetTotal()=1
	// issues.Issues[0].GetNumber()=49450
}
