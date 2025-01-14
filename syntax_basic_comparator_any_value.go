package jsonpath

import "encoding/json"

type syntaxBasicAnyValueComparator struct {
}

func (c *syntaxBasicAnyValueComparator) typeCast(values []interface{}) bool {
	var foundValue bool
	for index := range values {
		switch typedValue := values[index].(type) {
		case json.Number:
			foundValue = true
			values[index], _ = typedValue.Float64()
		case struct{}:
		default:
			foundValue = true
		}
	}
	return foundValue
}
