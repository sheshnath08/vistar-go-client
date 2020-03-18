package parameter

import "strconv"

func ParseStringValue(params map[string]interface{}, key string,
	def string, overrideValue string) string {
	if overrideValue != "" {
		return overrideValue
	}

	val, ok := params[key]
	if !ok {
		return def
	}
	return val.(string)
}

func ParseIntValue(params map[string]interface{}, key string, def int64,
	overrideValue int64) int64 {
	if overrideValue != int64(0) {
		return overrideValue
	}

	val, ok := params[key]
	if !ok {
		return def
	}
	// Number values are saved as float64.
	return int64(val.(float64))
}

func ParseFloatValue(params map[string]interface{}, key string,
	def float64, overrideValue float64) float64 {
	if overrideValue != float64(0) {
		return overrideValue
	}

	val, ok := params[key]
	if !ok {
		return def
	}
	return val.(float64)
}

func ParseBoolValue(params map[string]interface{}, key string, def bool,
	overrideValue bool) bool {
	if overrideValue {
		return overrideValue
	}

	val, ok := params[key]
	if !ok {
		return def
	}
	return val.(bool)
}

func ParseArrayValue(params map[string]interface{}, key string,
	def []string, overrideValue []string) []string {
	if len(overrideValue) > 0 {
		return overrideValue
	}

	val, ok := params[key]
	if !ok {
		return def
	}

	return val.([]string)
}

func ParseBoolScreenParam(params map[string]interface{}, key string,
	def bool) bool {
	sval, ok := params[key]
	if !ok {
		return def
	}
	val, err := strconv.ParseBool(sval.(string))
	if err != nil {
		return def
	}
	return val
}

func ParseIntScreenParam(params map[string]interface{}, key string,
	def int64) int64 {
	sval, ok := params[key]
	if !ok {
		return def
	}

	val, err := strconv.ParseInt(sval.(string), 10, 64)
	if err != nil {
		return def
	}
	return val
}
