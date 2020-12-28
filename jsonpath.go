package jsonpath

import (
	"regexp"
	"sync"
)

var parseMutex sync.Mutex
var parser = pegJSONPathParser{}
var unescapeRegex = regexp.MustCompile(`\\(.)`)

// Retrieve returns the retrieved JSON using the given JSONPath.
func Retrieve(jsonPath string, src interface{}, config ...Config) ([]interface{}, error) {
	jsonPathFunc, err := Parse(jsonPath, config...)
	if err != nil {
		return nil, err
	}
	return jsonPathFunc(src)
}

// Parse returns the parser function using the given JSONPath.
func Parse(jsonPath string, config ...Config) (func(src interface{}) ([]interface{}, error), error) {
	parseMutex.Lock()
	defer parseMutex.Unlock()

	parser.Buffer = jsonPath
	parser.jsonPathParser = jsonPathParser{
		unescapeRegex: unescapeRegex,
	}

	if parser.parse == nil {
		parser.Init()
	} else {
		parser.Reset()
	}

	if len(config) > 0 {
		parser.jsonPathParser.filterFunctions = config[0].filterFunctions
		parser.jsonPathParser.aggregateFunctions = config[0].aggregateFunctions
		parser.jsonPathParser.accessorMode = config[0].accessorMode
	}

	parser.Parse()
	parser.Execute()

	if parser.thisError != nil {
		return nil, parser.thisError
	}

	var parserFuncMutex sync.Mutex
	return func(src interface{}) ([]interface{}, error) {
		parserFuncMutex.Lock()
		defer parserFuncMutex.Unlock()

		result := make([]interface{}, 0)
		parser.jsonPathParser.srcJSON = &src
		parser.jsonPathParser.resultPtr = &result
		if err := parser.jsonPathParser.root.retrieve(src); err != nil {
			return nil, err
		}
		return result, nil
	}, nil
}
