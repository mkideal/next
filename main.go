//go:build !js && !wasm

package main

import (
	"os"

	"github.com/next/next/src/types"
)

func main() {
	exec(types.StandardPlatform(), os.Args)
}
