//go:build !js && !wasm

package main

import (
	"os"

	"github.com/next/next/src/compile"
)

func main() {
	exec(compile.StandardPlatform(), os.Args)
}
