![🚧 Under construction 👷‍♂️](https://i.imgur.com/LEP2R3N.png)

# go-actions-toolkit/core

✅ Get inputs, set outputs, add error annotations, and more basic GitHub Actions operations in Go

<table align=center><td>

```go
name, err := core.GetInput("name", &core.InputOptions{
	Required: ptr(true),
})
if err != nil {
	core.SetFailed(err)
}

favoriteColor, _ := core.GetInput("favorite-color", nil)
if favoriteColor == "" {
	favoriteColor = "rainbow"
}

log.Printf("Hello %s!\n", name)
log.Printf("I like %s too.\n", favoriteColor)

core.SetOutput("first-name", strings.Split(name, " ")[0])
```

</table>

🧰 Part of the GitHub Actions for Go toolkit \
👨‍💻 Lets you write your GitHub Actions code in Go \
🤩 Implements the complete [@actions/core](https://www.npmjs.com/package/@actions/core) API surface \
🚀 Works great with [jcbhmr/configure-executable-action](https://github.com/jcbhmr/configure-executable-action)

## Installation

You can install this Go package straight from GitHub using `go get`:

```sh
go get github.com/jcbhmr/go-actions-toolkit/core
```

Or you can `import` it in your code and run `go mod tidy`:

```go
import "github.com/jcbhmr/go-actions-toolkit/core"
```

## Usage

ℹ This Go module tries to mirror the API surface, feel, and functionality of the JavaScript [@actions/core](https://www.npmjs.com/package/@actions/core) package. See [Differences from @actions/core](#differences-from-actionscore) for more information.

The basic structure of a Go-based GitHub action is a `main()` function with a `defer`-ed `core.SetFailed()` call to show an error annotation whenever the action panics.

```go
package main

import (
	"github.com/jcbhmr/go-actions-toolkit/core"
)


func main() {
	defer func(){
		if err := recover(); err != nil {
			core.SetFailed(err)
		}
	}()

	// The rest of your code here! 🚀
}
```

⭐ **Try to use `core.SetFailed()`** so that you get a pretty error annotation in the summary page. `os.Exit(1)` doesn't show that.

<p align=center>
  <img src="https://gist.github.com/assets/61068799/1297827c-5a6b-4af5-ae91-b9076dc25cf5">
</p>

Inside your GitHub action code you'll probably be using some inputs and outputs. To define a required input make sure to pass a pointer to `true` in the `Required` field of the `core.InputOptions` struct.

```go
name, err := core.GetInput("name", &core.InputOptions{
	Required: ptr(true),
})
// `err` will be non-nil if the input is required and not provided.
// Use standard Go error handling to deal with it.
if err != nil {
	core.SetFailed(err)
}
```

<details><summary>📚 The <code>ptr()</code> function</summary>

This package uses `*T` types to represent optional parameters. Go doesn't support taking the `&` address of a literal. You need to use a `ptr()` or similar helper function or a defined variable to be able to turn a `T` into a `*T`.

```go
func ptr[T any](v T) *T {
	return &v
}

func main() {
	// ⛔ This doesn't compile. You can't take the address of a literal.
	// options1 := core.InputOptions{Required: &true}

	// 🤷‍♂️ This requires an additional temp var. Not ideal.
	required := true
	options2 := core.InputOptions{Required: &required}

	// 🤩 Much better! No need for an extra var.
	options3 := core.InputOptions{Required: ptr(true)}
}
```

</details>

If you have a non-required field it's safe to ignore the `err` return value. The only time `GetInput` will return an error is if the input is required and not provided.

```go
name, _ := core.GetInput("name", nil)
// `name` will still default to `""` if the input wasn't present.
if name == "" {
	name = "Ada Lovelace"
}
```

### Differences from @actions/core

<sup>Have a better way to do one of these compromises? [Tell me! ❤️](https://github.com/jcbhmr/go-actions-toolkit/issues/new)</sup>

- Named after to the actions/toolkit GitHub repository instead of after the @actions/* published packages.
- `core.ExitCode.Success` from @actions/core is now `core.ExitCodeSuccess` in Go
- `core.platform.isWindows` and other `core.platform.*` stuff from @actions/core is now `core.PlatformIsWindows` and `core.Platform*` in Go
- `core.SetFailed()` in Go **immediately exits the program** via `os.Exit(1)` instead of setting a future `process.exitCode = 1` like in TypeScript.
- This Go toolkit core module **does not** depend on the other toolkit http and exec modules. @actions/core does.
- In lieu of native Go optional parameters this module uses `*T` to represent optionals.
