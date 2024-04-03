package core_test

import (
	"fmt"
	"os"

	"github.com/jcbhmr/actions-toolkit.go/core"
)

func ptr[T any](v T) *T {
	return &v
}

func unwrap1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func catch(f func(any)) {
	if err := recover(); err != nil {
		f(err)
	}
}

func Example() {
	os.Setenv("INPUT_NAME", "Ada Lovelace")
	defer catch(core.SetFailed)
	name := unwrap1(core.GetInput("name", &core.InputOptions{Required: ptr(true)}))
	favoriteColor := unwrap1(core.GetInput("favorite-color", nil))
	if favoriteColor == "" {
		favoriteColor = "rainbow"
	}
	fmt.Printf("Hello %s! I like %s too.\n", name, favoriteColor)
	// Output: Hello Ada Lovelace! I like rainbow too.
}

func ExampleGetInput() {
	os.Setenv("INPUT_NAME", "Alan Turing")
	name := unwrap1(core.GetInput("name", &core.InputOptions{Required: ptr(true)}))
	fmt.Printf("Hello %s!", name)
	// Output: Hello Alan Turing!
}

func ExampleSetOutput() {
	core.SetOutput("message", "Hello world!")
	// Output: ::set-output name=message::Hello world!
}

func ExampleSetFailed() {
	defer catch(core.SetFailed)
	path := unwrap1(core.GetInput("path", &core.InputOptions{Required: ptr(true)}))
	fmt.Printf("Reading %s...\n", path)
}
