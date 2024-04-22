# go-actions-toolkit/github

<table align=center><td>

```go

```

</table>

## Installation

```sh
go get github.com/jcbhmr/go-actions-toolkit/github
```

## Usage

```go
package main

import (
    "fmt"

    "github.com/jcbhmr/go-actions-toolkit/github"
)

func main() {
    repo, _ := github.Context.Repo()
    fmt.Printf("This is the %s/%s repository", repo.Owner, repo.Repo)
    // Output: This is the jcbhmr/go-actions-toolkit repository
}
```

## Development