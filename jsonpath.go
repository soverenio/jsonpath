package jsonpath

import (
	"regexp"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return &pegJSONPathParser{}
	},
}
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
func Parse(jsonPath string, config ...Config) (f func(src interface{}) ([]interface{}, error), err error) {
	var parser = pool.Get().(*pegJSONPathParser)
	defer func() {
		if exception := recover(); exception != nil {
			if _err, ok := exception.(error); ok {
				err = _err
			}
		}
		parser.jsonPathParser = jsonPathParser{}
		pool.Put(parser)
	}()

	parser.Buffer = jsonPath

	if parser.parse == nil {
		parser.Init()
	} else {
		parser.Reset()
	}

	parser.jsonPathParser.unescapeRegex = unescapeRegex

	if len(config) > 0 {
		parser.jsonPathParser.filterFunctions = config[0].filterFunctions
		parser.jsonPathParser.aggregateFunctions = config[0].aggregateFunctions
		parser.jsonPathParser.accessorMode = config[0].accessorMode
	}

	parser.Parse()
	parser.Execute()

	root := parser.jsonPathParser.root
	return func(src interface{}) ([]interface{}, error) {
		container := getContainer()
		defer func() {
			putContainer(container)
		}()

		if err := root.retrieve(src, src, container); err != nil {
			return nil, err.(error)
		}

		result := make([]interface{}, len(container.result))
		for index := range result {
			result[index] = container.result[index]
		}

		return result, nil

	}, nil
}
