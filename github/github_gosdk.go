//go:build toolkitgithub_gosdk

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	nethttplibrary "github.com/microsoft/kiota-http-go"
	"github.com/octokit/go-sdk/pkg/authentication"
	gosdkgithub "github.com/octokit/go-sdk/pkg/github"
)

func GetApiClient(token string) (*gosdkgithub.ApiClient, error) {
	provider := authentication.NewTokenProvider(
		authentication.WithAuthorizationToken(token),
		authentication.WithUserAgent("github.com/jcbhmr/go-actions-toolkit/github"),
	)
	adapter, err := nethttplibrary.NewNetHttpRequestAdapter(provider)
	if err != nil {
		return nil, err
	}
	apiClient := gosdkgithub.NewApiClient(adapter)
	return apiClient, nil
}