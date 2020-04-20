package parameter

import (
	"fmt"
	"strconv"
)

func ParseStringValue(params map[string]interface{}, key string,
	def string, overrideValue string) (string, error) {
	if overrideValue != "" {
		return overrideValue, nil
	}

	val, ok := params[key]
	if !ok {
		return def, nil
	}

	v, ok := val.(string)
	if !ok {
		return "", fmt.Errorf(
			"Invalid typed value for param %s: %v, should be of type string",
			key, val)
	}
	return v, nil
}

func ParseIntValue(params map[string]interface{}, key string, def int64,
	overrideValue int64) (int64, error) {
	if overrideValue != int64(0) {
		return overrideValue, nil
	}

	val, ok := params[key]
	if !ok {
		return def, nil
	}
	// Number values are saved as float64.
	v, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf(
			"Invalid typed value for param %s: %v, should be of type number",
			key, val)
	}

	return int64(v), nil
}

func ParseFloatValue(params map[string]interface{}, key string,
	def float64, overrideValue float64) (float64, error) {
	if overrideValue != float64(0) {
		return overrideValue, nil
	}

	val, ok := params[key]
	if !ok {
		return def, nil
	}

	v, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf(
			"Invalid typed value for param %s: %v, should be of type number",
			key, val)
	}

	return v, nil
}

func ParseBoolValue(params map[string]interface{}, key string, def bool,
	overrideValue bool) (bool, error) {
	if overrideValue {
		return overrideValue, nil
	}

	val, ok := params[key]
	if !ok {
		return def, nil
	}

	v, ok := val.(bool)

	if !ok {
		return false, fmt.Errorf(
			"Invalid typed value for param %s: %v, should be of type boolean",
			key, val)
	}

	return v, nil
}

func ParseArrayValue(params map[string]interface{}, key string,
	def []string, overrideValue []string) ([]string, error) {
	if len(overrideValue) > 0 {
		return overrideValue, nil
	}

	val, ok := params[key]
	if !ok {
		return def, nil
	}

	v, ok := val.([]string)
	if !ok {
		return nil, fmt.Errorf("Invalid typed value for param %s: %v, "+
			"should be of type list of string", key, val)
	}

	return v, nil
}

func ParseBoolScreenParam(params map[string]interface{}, key string,
	def bool) (bool, error) {
	sval, ok := params[key]
	if !ok {
		return def, nil
	}

	val, err := strconv.ParseBool(sval.(string))
	if err != nil {
		return false, err
	}

	return val, nil
}

func ParseIntScreenParam(params map[string]interface{}, key string,
	def int64) (int64, error) {
	sval, ok := params[key]
	if !ok {
		return def, nil
	}

	val, err := strconv.ParseInt(sval.(string), 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
