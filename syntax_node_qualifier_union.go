package jsonpath

import "reflect"

type syntaxUnionQualifier struct {
	*syntaxBasicNode

	subscripts []syntaxSubscript
}

func (u *syntaxUnionQualifier) retrieve(
	root, current interface{}, result *[]interface{}) error {

	srcArray, ok := current.([]interface{})
	if !ok {
		foundType := `null`
		if current != nil {
			foundType = reflect.TypeOf(current).String()
		}
		return ErrorTypeUnmatched{
			expectedType: `array`,
			foundType:    foundType,
			path:         u.text,
		}
	}

	var resultIndexes []int
	for index := range u.subscripts {
		resultIndexes = append(resultIndexes, u.subscripts[index].getIndexes(srcArray)...)
	}

	if u.isValueGroup() {
		childErrorMap := make(map[error]struct{}, 1)
		var lastError error
		for index := range resultIndexes {
			if err := u.retrieveListNext(root, srcArray, resultIndexes[index], result); err != nil {
				childErrorMap[err] = struct{}{}
				lastError = err
			}
		}

		if len(*result) == 0 {
			switch len(childErrorMap) {
			case 0:
				return ErrorNoneMatched{path: u.text}
			case 1:
				return lastError
			default:
				return ErrorNoneMatched{path: u.next.getConnectedText()}
			}
		}

		return nil
	}

	if len(resultIndexes) == 0 {
		return ErrorIndexOutOfRange{path: u.text}
	}

	return u.retrieveListNext(root, srcArray, resultIndexes[0], result)
}

func (u *syntaxUnionQualifier) merge(union *syntaxUnionQualifier) {
	u.subscripts = append(u.subscripts, union.subscripts...)
}
