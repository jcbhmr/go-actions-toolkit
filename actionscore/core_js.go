package core

import (
	"sync"
	"syscall/js"

	"github.com/jcbhmr/go-toolkit/actionscore/internal/utils"
)

func catchJSError(err *error) {
	if r := recover(); r != nil {
		if jsErr, ok := r.(js.Error); ok {
			*err = jsErr
			if jsErr.Value.Type() == js.TypeObject || jsErr.Value.Type() == js.TypeFunction {
				if name := jsErr.Value.Get("name"); name.Type() == js.TypeString && name.String() == "TypeError" {
					panic(jsErr)
				}
			}
		} else {
			panic(r)
		}
	}
}

func await2(promise js.Value) (v js.Value, err error) {
	promise = js.Global().Get("Promise").Call("resolve", promise)
	var wg sync.WaitGroup
	resolveCallback := js.FuncOf(func(this js.Value, args []js.Value) any {
		defer wg.Done()
		v = args[0]
		return js.Undefined()
	})
	defer resolveCallback.Release()
	rejectCallback := js.FuncOf(func(this js.Value, args []js.Value) any {
		defer wg.Done()
		err = js.Error{Value: args[0]}
		return js.Undefined()
	})
	defer rejectCallback.Release()
	wg.Add(1)
	promise.Call("then", resolveCallback, rejectCallback)
	wg.Wait()
	return
}

func asyncFuncOf2(f func(this js.Value, args []js.Value) (any, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		cb := js.FuncOf(func(_ js.Value, args2 []js.Value) any {
			go func() {
				r, err := f(this, args)
				if err != nil {
					args2[1].Invoke(js.Global().Get("Error").New(err.Error()))
				} else {
					args2[0].Invoke(r)
				}
			}()
			return js.Undefined()
		})
		defer cb.Release()
		return js.Global().Get("Promise").New(cb)
	})
}

var core = js.Global().Get("@actions/core")

func exportVariable(name string, val any) (err error) {
	defer catchJSError(&err)
	// Need to lower Go structs to JSON strings. Reuse logic from utils.
	val2, err := utils.ToCommandValue(val)
	if err != nil {
		return err
	}
	core.Get("exportVariable").Invoke(name, val2)
	return
}

func setSecret(secret string) (err error) {
	defer catchJSError(&err)
	core.Get("setSecret").Invoke(secret)
	return
}

func addPath(inputPath string) (err error) {
	defer catchJSError(&err)
	core.Get("addPath").Invoke(inputPath)
	return
}

func getInput(name string, options *InputOptions) (_ string, err error) {
	defer catchJSError(&err)
	if options == nil {
		return core.Get("getInput").Invoke(name).String(), nil
	} else {
		options2 := map[string]any{}
		if options.Required != nil {
			options2["required"] = *options.Required
		}
		if options.TrimWhitespace != nil {
			options2["trimWhitespace"] = *options.TrimWhitespace
		}
		v := core.Get("getInput").Invoke(name, options2)
		if v.Type() != js.TypeString {
			panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
		}
		return v.String(), nil
	}
}

func getMultilineInput(name string, options *InputOptions) (_ []string, err error) {
	defer catchJSError(&err)
	if options == nil {
		v := core.Get("getMultilineInput").Invoke(name)
		v2 := make([]string, v.Length())
		for i := 0; i < v.Length(); i++ {
			v2Raw := v.Index(i)
			if v2Raw.Type() != js.TypeString {
				panic(&js.ValueError{Method: "Value.String", Type: v2Raw.Type()})
			}
			v2[i] = v2Raw.String()
		}
		return v2, nil
	} else {
		options2 := map[string]any{}
		if options.Required != nil {
			options2["required"] = *options.Required
		}
		if options.TrimWhitespace != nil {
			options2["trimWhitespace"] = *options.TrimWhitespace
		}
		v := core.Get("getMultilineInput").Invoke(name, options2)
		v2 := make([]string, v.Length())
		for i := 0; i < v.Length(); i++ {
			v2Raw := v.Index(i)
			if v2Raw.Type() != js.TypeString {
				panic(&js.ValueError{Method: "Value.String", Type: v2Raw.Type()})
			}
			v2[i] = v2Raw.String()
		}
		return v2, nil
	}
}

func getBooleanInput(name string, options *InputOptions) (_ bool, err error) {
	defer catchJSError(&err)
	if options == nil {
		return core.Get("getBooleanInput").Invoke(name).Bool(), nil
	} else {
		options2 := map[string]any{}
		if options.Required != nil {
			options2["required"] = *options.Required
		}
		if options.TrimWhitespace != nil {
			options2["trimWhitespace"] = *options.TrimWhitespace
		}
		return core.Get("getBooleanInput").Invoke(name, options2).Bool(), nil
	}
}

func setOutput(name string, value any) (err error) {
	defer catchJSError(&err)
	// Need to lower Go structs to JSON strings. Reuse logic from utils.
	value2, err := utils.ToCommandValue(value)
	if err != nil {
		return err
	}
	core.Get("setOutput").Invoke(name, value2)
	return
}

func setCommandEcho(enabled bool) (err error) {
	defer catchJSError(&err)
	core.Get("setCommandEcho").Invoke(enabled)
	return
}

func setFailed(message string) (err error) {
	defer catchJSError(&err)
	core.Get("setFailed").Invoke(message)
	return
}

func isDebug() bool {
	return core.Get("isDebug").Invoke().Bool()
}

func debug(message string) (err error) {
	defer catchJSError(&err)
	core.Get("debug").Invoke(message)
	return
}

func errorFunc(message string, properties *AnnotationProperties) (err error) {
	defer catchJSError(&err)
	if properties == nil {
		core.Get("error").Invoke(message)
	} else {
		properties2 := map[string]any{}
		if properties.Title != nil {
			properties2["title"] = *properties.Title
		}
		if properties.File != nil {
			properties2["file"] = *properties.File
		}
		if properties.StartLine != nil {
			properties2["startLine"] = *properties.StartLine
		}
		if properties.EndLine != nil {
			properties2["endLine"] = *properties.EndLine
		}
		if properties.StartColumn != nil {
			properties2["startColumn"] = *properties.StartColumn
		}
		if properties.EndColumn != nil {
			properties2["endColumn"] = *properties.EndColumn
		}
		core.Get("error").Invoke(message, properties2)
	}
	return
}

func warning(message string, properties *AnnotationProperties) (err error) {
	defer catchJSError(&err)
	if properties == nil {
		core.Get("warning").Invoke(message)
	} else {
		properties2 := map[string]any{}
		if properties.Title != nil {
			properties2["title"] = *properties.Title
		}
		if properties.File != nil {
			properties2["file"] = *properties.File
		}
		if properties.StartLine != nil {
			properties2["startLine"] = *properties.StartLine
		}
		if properties.EndLine != nil {
			properties2["endLine"] = *properties.EndLine
		}
		if properties.StartColumn != nil {
			properties2["startColumn"] = *properties.StartColumn
		}
		if properties.EndColumn != nil {
			properties2["endColumn"] = *properties.EndColumn
		}
		core.Get("warning").Invoke(message, properties2)
	}
	return
}

func notice(message string, properties *AnnotationProperties) (err error) {
	defer catchJSError(&err)
	if properties == nil {
		core.Get("notice").Invoke(message)
	} else {
		properties2 := map[string]any{}
		if properties.Title != nil {
			properties2["title"] = *properties.Title
		}
		if properties.File != nil {
			properties2["file"] = *properties.File
		}
		if properties.StartLine != nil {
			properties2["startLine"] = *properties.StartLine
		}
		if properties.EndLine != nil {
			properties2["endLine"] = *properties.EndLine
		}
		if properties.StartColumn != nil {
			properties2["startColumn"] = *properties.StartColumn
		}
		if properties.EndColumn != nil {
			properties2["endColumn"] = *properties.EndColumn
		}
		core.Get("notice").Invoke(message, properties2)
	}
	return
}

func info(message string) (err error) {
	defer catchJSError(&err)
	core.Get("info").Invoke(message)
	return
}

func startGroup(name string) (err error) {
	defer catchJSError(&err)
	core.Get("startGroup").Invoke(name)
	return
}

func endGroup() (err error) {
	defer catchJSError(&err)
	core.Get("endGroup").Invoke()
	return
}

func group(name string, fn func() (any, error)) (v any, err error) {
	defer catchJSError(&err)
	asyncFn := asyncFuncOf2(func(this js.Value, args []js.Value) (any, error) {
		return fn()
	})
	defer asyncFn.Release()
	v, err = await2(core.Get("group").Invoke(name, asyncFn))
	return
}

func saveState(name string, value any) (err error) {
	defer catchJSError(&err)
	value2, err := utils.ToCommandValue(value)
	if err != nil {
		return err
	}
	core.Get("saveState").Invoke(name, value2)
	return
}

func getState(name string) string {
	v := core.Get("getState").Invoke(name)
	if v.Type() != js.TypeString {
		panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
	}
	return v.String()
}

func getIDToken(aud *string) (string, error) {
	var p js.Value
	if aud == nil {
		p = core.Get("getIDToken").Invoke()
	} else {
		p = core.Get("getIDToken").Invoke(*aud)
	}
	v, err := await2(p)
	if err != nil {
		return "", err
	}
	if v.Type() != js.TypeString {
		panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
	}
	return v.String(), nil
}

func toPosixPath(p string) string {
	v := core.Get("toPosixPath").Invoke(p)
	if v.Type() != js.TypeString {
		panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
	}
	return v.String()
}

func toWin32Path(p string) string {
	v := core.Get("toWin32Path").Invoke(p)
	if v.Type() != js.TypeString {
		panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
	}
	return v.String()
}

func toPlatformPath(p string) string {
	v := core.Get("toPlatformPath").Invoke(p)
	if v.Type() != js.TypeString {
		panic(&js.ValueError{Method: "Value.String", Type: v.Type()})
	}
	return v.String()
}

type summaryType struct{ value js.Value }

var summary = summaryType{core.Get("summary")}

var markdownSummary = summaryType{core.Get("markdownSummary")}
