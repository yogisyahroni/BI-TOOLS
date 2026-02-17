package formula_engine

import (
	"fmt"
	"strings"
)

func funcUpper(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("UPPER requires 1 argument")
	}
	return strings.ToUpper(fmt.Sprintf("%v", args[0])), nil
}

func funcLower(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("LOWER requires 1 argument")
	}
	return strings.ToLower(fmt.Sprintf("%v", args[0])), nil
}

func funcConcat(args []interface{}) (interface{}, error) {
	var sb strings.Builder
	for _, arg := range args {
		if subArgs, ok := arg.([]interface{}); ok {
			for _, subArg := range subArgs {
				sb.WriteString(fmt.Sprintf("%v", subArg))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%v", arg))
		}
	}
	return sb.String(), nil
}

func funcLen(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("LEN requires 1 argument")
	}
	return float64(len(fmt.Sprintf("%v", args[0]))), nil
}

func funcTrim(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TRIM requires 1 argument")
	}
	return strings.TrimSpace(fmt.Sprintf("%v", args[0])), nil
}

func funcLeft(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("LEFT requires 2 arguments")
	}
	str := fmt.Sprintf("%v", args[0])
	n, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}
	if int(n) >= len(str) {
		return str, nil
	}
	if n < 0 {
		return "", fmt.Errorf("length cannot be negative")
	}
	return str[:int(n)], nil
}

func funcRight(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("RIGHT requires 2 arguments")
	}
	str := fmt.Sprintf("%v", args[0])
	n, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}
	if int(n) >= len(str) {
		return str, nil
	}
	if n < 0 {
		return "", fmt.Errorf("length cannot be negative")
	}
	return str[len(str)-int(n):], nil
}
