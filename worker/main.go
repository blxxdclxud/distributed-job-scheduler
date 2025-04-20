package main

import (
	"fmt"

	"github.com/Shopify/go-lua"
)

const testProgram string = `
a = 1 + 2
b = 3 * 4
return a * b
`

func main() {
	fmt.Println("worker")
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.LoadString(l, testProgram); err != nil {
		fmt.Println("Error loading program:", err)
		return
	}

	if err := l.ProtectedCall(0, 1, 0); err != nil {
		fmt.Println("Error executing program:", err)
		return
	}

	fmt.Println("Result:", l.ToValue(0))
}
