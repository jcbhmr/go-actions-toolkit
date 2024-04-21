package core_test

import (
	"fmt"
	"os"

	"github.com/jcbhmr/go-actions-toolkit/core"
)

func ptr[T any](v T) *T {
	return &v
}

func Example() {
	// Demo purposes only.
	os.Setenv("INPUT_NAME", "Ada Lovelace")
	// os.Setenv("INPUT_FAVORITE-COLOR", "blue")

	defer func() {
		if err := recover(); err != nil {
			core.SetFailed(err)
		}
	}()
	name, err := core.GetInput("name", &core.InputOptions{Required: ptr(true)})
	if err != nil {
		core.SetFailed(err)
	}
	favoriteColor, _ := core.GetInput("favorite-color", nil)
	if favoriteColor == "" {
		favoriteColor = "rainbow"
	}
	fmt.Printf("Hello %s! I like %s too.\n", name, favoriteColor)
	// Output: Hello Ada Lovelace! I like rainbow too.
}

func ExampleGetInput() {
	// Demo purposes only.
	os.Setenv("INPUT_NAME", "Alan Turing")

	name, err := core.GetInput("name", &core.InputOptions{Required: ptr(true)})
	if err != nil {
		core.SetFailed(err)
	}
	fmt.Printf("Hello %s!\n", name)
	// Output: Hello Alan Turing!
}

func ExampleSetOutput() {
	// Will use `GITHUB_OUTPUT` path if set.
	core.SetOutput("message", "Hello world!")
	// Output: ::set-output name=message::Hello world!
}
