package utils

import (
	"encoding/json"
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

	// Inject polyfills
	_, err := vm.RunString(polyfills)
	if err != nil {
		return "", fmt.Errorf("polyfill injection error: %w", err)
	}

	// Execute the script to load definitions
	_, err = vm.RunString(scriptContent)
	if err != nil {
		return "", fmt.Errorf("script compilation error: %w", err)
	}

	// Get the main function
	mainFn, ok := goja.AssertFunction(vm.Get("subMod"))
	if !ok {
		return "", fmt.Errorf("subMod function not found in script")
	}

	// Call the main function
	result, err := mainFn(goja.Undefined(), vm.ToValue(input), vm.ToValue(clientType))
	if err != nil {
		return "", fmt.Errorf("script execution error: %w", err)
	}

	return result.String(), nil
}

// RunNodeFilterScript executes a JavaScript script to filter nodes.
// The script is expected to define a function `filterNode(nodes, clientType)` that returns a modified nodes array.
func RunNodeFilterScript(scriptContent string, nodesJSON []byte, clientType string) ([]byte, error) {
	vm := goja.New()

	// Inject console object
	vm.Set("console", map[string]interface{}{
		"log":   fmt.Println,
		"info":  fmt.Println,
		"warn":  fmt.Println,
		"error": fmt.Println,
	})

	// Inject polyfills
	_, err := vm.RunString(polyfills)
	if err != nil {
		return nil, fmt.Errorf("polyfill injection error: %w", err)
	}

	// Execute the script to load definitions
	_, err = vm.RunString(scriptContent)
	if err != nil {
		return nil, fmt.Errorf("script compilation error: %w", err)
	}

	// Get the filterNode function
	filterFn, ok := goja.AssertFunction(vm.Get("filterNode"))
	if !ok {
		// If function not found, return original nodes (or error? Plan said error, but maybe better to just return nil error and original nodes if we want to be lenient.
		// However, explicit error is better for debugging.
		return nil, fmt.Errorf("filterNode function not found in script")
	}

	// Unmarshal nodes
	var nodes interface{}
	if err := json.Unmarshal(nodesJSON, &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	// Call the function
	result, err := filterFn(goja.Undefined(), vm.ToValue(nodes), vm.ToValue(clientType))
	if err != nil {
		return nil, fmt.Errorf("script execution error: %w", err)
	}

	// Marshal result back to JSON
	// The result should be the modified nodes array
	resNodes := result.Export()
	newJSON, err := json.Marshal(resNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return newJSON, nil
}

const polyfills = `
if (!String.prototype.includes) {
  String.prototype.includes = function(search, start) {
    if (typeof start !== 'number') {
      start = 0;
    }
    if (start + search.length > this.length) {
      return false;
    } else {
      return this.indexOf(search, start) !== -1;
    }
  };
}

if (!String.prototype.startsWith) {
  String.prototype.startsWith = function(searchString, position) {
    position = position || 0;
    return this.substr(position, searchString.length) === searchString;
  };
}

if (!String.prototype.endsWith) {
  String.prototype.endsWith = function(searchString, position) {
    var subjectString = this.toString();
    if (typeof position !== 'number' || !isFinite(position) || Math.floor(position) !== position || position > subjectString.length) {
      position = subjectString.length;
    }
    position -= searchString.length;
    var lastIndex = subjectString.lastIndexOf(searchString, position);
    return lastIndex !== -1 && lastIndex === position;
  };
}

if (!Array.prototype.find) {
  Object.defineProperty(Array.prototype, 'find', {
    value: function(predicate) {
      if (this == null) {
        throw new TypeError('"this" is null or not defined');
      }
      var o = Object(this);
      var len = o.length >>> 0;
      if (typeof predicate !== 'function') {
        throw new TypeError('predicate must be a function');
      }
      var thisArg = arguments[1];
      var k = 0;
      while (k < len) {
        var kValue = o[k];
        if (predicate.call(thisArg, kValue, k, o)) {
          return kValue;
        }
        k++;
      }
      return undefined;
    },
    configurable: true,
    writable: true
  });
}

if (!Array.prototype.findIndex) {
  Object.defineProperty(Array.prototype, 'findIndex', {
    value: function(predicate) {
      if (this == null) {
        throw new TypeError('"this" is null or not defined');
      }
      var o = Object(this);
      var len = o.length >>> 0;
      if (typeof predicate !== 'function') {
        throw new TypeError('predicate must be a function');
      }
      var thisArg = arguments[1];
      var k = 0;
      while (k < len) {
        var kValue = o[k];
        if (predicate.call(thisArg, kValue, k, o)) {
          return k;
        }
        k++;
      }
      return -1;
    },
    configurable: true,
    writable: true
  });
}

if (!Array.prototype.includes) {
  Object.defineProperty(Array.prototype, 'includes', {
    value: function(searchElement, fromIndex) {
      if (this == null) {
        throw new TypeError('"this" is null or not defined');
      }
      var o = Object(this);
      var len = o.length >>> 0;
      if (len === 0) {
        return false;
      }
      var n = fromIndex | 0;
      var k = Math.max(n >= 0 ? n : len - Math.abs(n), 0);
      while (k < len) {
        if (o[k] === searchElement) {
          return true;
        }
        k++;
      }
      return false;
    }
  });
}

if (typeof Object.assign != 'function') {
  Object.assign = function(target) {
    'use strict';
    if (target == null) {
      throw new TypeError('Cannot convert undefined or null to object');
    }
    target = Object(target);
    for (var index = 1; index < arguments.length; index++) {
      var source = arguments[index];
      if (source != null) {
        for (var key in source) {
          if (Object.prototype.hasOwnProperty.call(source, key)) {
            target[key] = source[key];
          }
        }
      }
    }
    return target;
  };
}
`
