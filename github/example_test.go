package github_test

import (
	"context"
	"fmt"
	"os"

	"github.com/jcbhmr/go-actions-toolkit/github"
)

func Example() {
	token := os.Getenv("GITHUB_TOKEN")
	client := unwrap1(github.GetGoGithubClient(token))
	repo := github.Context.Repo()
	info, _ := unwrap2(client.Repositories.Get(context.Background(), repo.Owner, repo.Repo))
	fmt.Printf("info.GetFullName()=%s\n", info.GetFullName())
	fmt.Printf("info.GetDescription()=%s\n", info.GetDescription())
	// Output:
	// info.GetFullName()=jcbhmr/actions-toolkit.go
	// info.GetDescription()=🐿️ GitHub Actions toolkit for your Go-based GitHub Actions
}

func ExampleGetGoGithubClient() {
	token := os.Getenv("GITHUB_TOKEN")
	client := unwrap1(github.GetGoGithubClient(token))
	query := "is:issue repo:nodejs/node is:open Proposal for single-mode packages with optional fallbacks for older versions of node"
	issues, _ := unwrap2(client.Search.Issues(context.Background(), query, nil))
	fmt.Printf("issues.GetTotal()=%d\n", issues.GetTotal())
	fmt.Printf("issues.Issues[0].GetNumber()=%d\n", issues.Issues[0].GetNumber())
	// Output:
	// issues.GetTotal()=1
	// issues.Issues[0].GetNumber()=49450
}
