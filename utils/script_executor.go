package utils

import (
	"fmt"
	"github.com/dop251/goja"
)

// RunScript executes a JavaScript script with the given input and client type.
// The script is expected to define a function `main(node, clientType)` that returns a string.
func RunScript(scriptContent string, input string, clientType string) (string, error) {
	vm := goja.New()

	// Inject console object
	vm.Set("console", map[string]interface{}{
		"log":   fmt.Println,
		"info":  fmt.Println,
		"warn":  fmt.Println,
		"error": fmt.Println,
	})

	// Execute the script to load definitions
	_, err := vm.RunString(scriptContent)
	if err != nil {
		return "", fmt.Errorf("script compilation error: %w", err)
	}

	// Get the main function
	mainFn, ok := goja.AssertFunction(vm.Get("main"))
	if !ok {
		return "", fmt.Errorf("main function not found in script")
	}

	// Call the main function
	result, err := mainFn(goja.Undefined(), vm.ToValue(input), vm.ToValue(clientType))
	if err != nil {
		return "", fmt.Errorf("script execution error: %w", err)
	}

	return result.String(), nil
}
