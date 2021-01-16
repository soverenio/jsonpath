package jsonpath

type syntaxFilterQualifier struct {
	*syntaxBasicNode

	query syntaxQuery
}

func (f *syntaxFilterQualifier) retrieve(
	root, current interface{}, container *bufferContainer) error {

	childErrorMap := make(map[error]struct{}, 1)
	var lastError error

	switch typedNodes := current.(type) {
	case map[string]interface{}:
		lastError = f.retrieveMap(root, typedNodes, container, childErrorMap)

	case []interface{}:
		lastError = f.retrieveList(root, typedNodes, container, childErrorMap)

	}

	if len(container.result) == 0 {
		switch len(childErrorMap) {
		case 0:
			return ErrorMemberNotExist{path: f.text}
		case 1:
			return lastError
		default:
			return ErrorNoneMatched{path: f.next.getConnectedText()}
		}
	}

	return nil
}

func (f *syntaxFilterQualifier) retrieveMap(
	root interface{}, srcMap map[string]interface{}, container *bufferContainer,
	childErrorMap map[error]struct{}) error {

	var lastError error

	sortKeys := container.getSortedKeys(srcMap)

	valueList := make([]interface{}, len(*sortKeys))
	for index := range *sortKeys {
		valueList[index] = srcMap[(*sortKeys)[index]]
	}

	valueList = f.query.compute(root, valueList, container)

	isEachResult := len(valueList) == len(srcMap)

	var nodeNotFound bool
	if !isEachResult {
		_, nodeNotFound = valueList[0].(struct{})
		if nodeNotFound {
			return nil
		}
	}

	for index := range *sortKeys {
		if isEachResult {
			_, nodeNotFound = valueList[index].(struct{})
		}
		if nodeNotFound {
			continue
		}
		if err := f.retrieveMapNext(root, srcMap, (*sortKeys)[index], container); err != nil {
			childErrorMap[err] = struct{}{}
			lastError = err
		}
	}

	container.putSortSlice(sortKeys)

	return lastError
}

func (f *syntaxFilterQualifier) retrieveList(
	root interface{}, srcList []interface{}, container *bufferContainer,
	childErrorMap map[error]struct{}) error {

	var lastError error

	valueList := f.query.compute(root, srcList, container)

	isEachResult := len(valueList) == len(srcList)

	var nodeNotFound bool
	if !isEachResult {
		_, nodeNotFound = valueList[0].(struct{})
		if nodeNotFound {
			return nil
		}
	}

	for index := range srcList {
		if isEachResult {
			_, nodeNotFound = valueList[index].(struct{})
		}
		if nodeNotFound {
			continue
		}
		if err := f.retrieveListNext(root, srcList, index, container); err != nil {
			childErrorMap[err] = struct{}{}
			lastError = err
		}
	}

	return lastError
}
