//go:build js && wasm

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"runtime/debug"
	"syscall/js"
)

type JSFunc = func(this js.Value, args []js.Value) any

func recoverJSFunc(fn JSFunc) JSFunc {
	return func(this js.Value, args []js.Value) (result any) {
		defer func() {
			if e := recover(); e != nil {
				result = js.Global().Get("Error").New(fmt.Sprintf("%s\n%s", e, debug.Stack()))
			}
		}()
		return fn(this, args)
	}
}

func main() {
	fmt.Println("next.wasm running")

	exit = func(code int) {
		panic(code)
	}
	usage = func(flagSet *flag.FlagSet, _ io.Writer, format string, args ...any) {
		if format != "" {
			panic(fmt.Sprintf(format, args...))
		}
		panic(flagSet.Name() + " help")
	}

	done := make(chan struct{})
	var w = js.Global()
	for name, fn := range map[string]JSFunc{
		"jsbNext": jsbNext,
	} {
		w.Set(name, js.FuncOf(recoverJSFunc(fn)))
	}
	<-done
	fmt.Println("next.wasm done")
}

func jsbNext(_ js.Value, params []js.Value) any {
	fmt.Println("executing jsbNext")
	defer fmt.Println("done jsbNext")
	args := make([]string, 0, len(params)+1)
	args = append(args, "next")
	for _, param := range params {
		args = append(args, param.String())
	}
	exec(jsPlatform{}, args)
	return nil
}

const (
	jsm_os            = "jsm_os_"
	jsm_os_getenv     = jsm_os + "getenv"
	jsm_os_readFile   = jsm_os + "readFile"
	jsm_os_writeFile  = jsm_os + "writeFile"
	jsm_os_isExist    = jsm_os + "isExist"
	jsm_os_isNotExist = jsm_os + "isNotExist"
)

type jsPlatform struct{}

func (jsPlatform) Getenv(key string) string {
	return js.Global().Call(jsm_os_getenv, key).String()
}

func (jsPlatform) UserHomeDir() (string, error) {
	return "", nil
}

func (jsPlatform) Stdin() io.Reader {
	return nil
}

func (jsPlatform) Stderr() io.Writer {
	return io.Discard
}

func (jsPlatform) ReadFile(filename string) ([]byte, error) {
	content := js.Global().Call(jsm_os_readFile, filename).String()
	return base64.StdEncoding.DecodeString(content)
}

func (jsPlatform) WriteFile(filename string, data []byte) error {
	content := base64.StdEncoding.EncodeToString(data)
	js.Global().Call(jsm_os_writeFile, filename, content)
	return nil
}

func (jsPlatform) IsExist(filename string) bool {
	return js.Global().Call(jsm_os_isExist, filename).Bool()
}

func (jsPlatform) IsNotExist(filename string) bool {
	return js.Global().Call(jsm_os_isNotExist, filename).Bool()
}
