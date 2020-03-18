package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStringValue(t *testing.T) {
	params := map[string]interface{}{
		"key": "test",
	}

	// Valid key with no override value, return value.
	assert.Equal(t, ParseStringValue(params, "key", "defaul-value", ""), "test")

	// Invalid key with no override value, return default value.
	assert.Equal(t, ParseStringValue(params, "invaldKey", "def", ""), "def")

	// Valid key with override value, return override value.
	assert.Equal(t, ParseStringValue(params, "key", "def", "override-value"),
		"override-value")

	// Invalid key with override value, return override value.
	assert.Equal(t, ParseStringValue(params, "invalid-key", "def",
		"override-value"), "override-value")

}

func TestParseIntValue(t *testing.T) {
	params := map[string]interface{}{
		"key": float64(100),
	}

	// Valid key with no override value, return value.
	assert.Equal(t, ParseIntValue(params, "key", int64(200), int64(0)),
		int64(100))

	// Invalid key with no override value, return default value.
	assert.Equal(t, ParseIntValue(params, "invaldKey", int64(200), int64(0)),
		int64(200))

	// Valid key with override value, return override value.
	assert.Equal(t, ParseIntValue(params, "key", int64(200), int64(300)),
		int64(300))

	// Invalid key with override value, return override value.
	assert.Equal(t, ParseIntValue(params, "invalidKey", int64(200), int64(0)),
		int64(200))
}

func TestParseFloatValue(t *testing.T) {
	params := map[string]interface{}{
		"key": float64(100),
	}

	// Valid key with no override value, return value.
	assert.Equal(t, ParseFloatValue(params, "key", float64(200), float64(0)),
		float64(100))

	// Invalid key with no override value, return default value.
	assert.Equal(t, ParseFloatValue(params, "invaldKey", float64(200),
		float64(0)), float64(200))

	// Valid key with override value, return override value.
	assert.Equal(t, ParseFloatValue(params, "key", float64(200), float64(300)),
		float64(300))

	// Invalid key with override value, return override value.
	assert.Equal(t, ParseFloatValue(params, "invalidKey", float64(200),
		float64(0)), float64(200))
}

func TestParseBoolValue(t *testing.T) {
	params := map[string]interface{}{
		"trueValueKey":  true,
		"falseValueKey": false,
	}

	// Valid key with no override value, return value.
	assert.Equal(t, ParseBoolValue(params, "trueValueKey", false, false), true)
	assert.Equal(t, ParseBoolValue(params, "falseValueKey", true, false), false)

	// Invalid key with no override value, return default value.
	assert.Equal(t, ParseBoolValue(params, "invaldKey", true, false), true)

	// Valid key with override value, return override value.
	assert.Equal(t, ParseBoolValue(params, "falseValueKey", false, true), true)

	// Invalid key with override value, return override value.
	assert.Equal(t, ParseBoolValue(params, "falseValueKey", false, true), true)
}

func TestParseArrayValue(t *testing.T) {
	params := map[string]interface{}{
		"key": []string{"test-1", "test-2"},
	}

	defaultValue := []string{"default-1", "default-2"}
	overrideValue := []string{"override-1", "override-2"}

	// Valid key with no override value, return value.
	assert.Equal(t, ParseArrayValue(params, "key", defaultValue, nil),
		[]string{"test-1", "test-2"})

	// Invalid key with no override value, return default value.
	assert.Equal(t, ParseArrayValue(params, "invaldKey", defaultValue, nil),
		[]string{"default-1", "default-2"})

	// Valid key with override value, return override value.
	assert.Equal(t, ParseArrayValue(params, "key", defaultValue,
		overrideValue), []string{"override-1", "override-2"})

	// Invalid key with override value, return override value.
	assert.Equal(t, ParseArrayValue(params, "invalidKey", defaultValue,
		overrideValue), []string{"override-1", "override-2"})
}

func TestParseBoolScreenParam(t *testing.T) {
	params := map[string]interface{}{
		"trueValueKey":  "true",
		"falseValueKey": "false",
		"key":           "invalid-bool",
	}

	// Valid key, return value.
	assert.Equal(t, ParseBoolScreenParam(params, "trueValueKey", false), true)
	assert.Equal(t, ParseBoolScreenParam(params, "falseValueKey", true), false)

	// Invalid key, return default value.
	assert.Equal(t, ParseBoolScreenParam(params, "invaldKey", true), true)

	// Valid key with invalid bool value, return default value.
	assert.Equal(t, ParseBoolScreenParam(params, "key", false), false)
}

func TestParseIntScreenParam(t *testing.T) {
	params := map[string]interface{}{
		"key":           "100",
		"invalidIntKey": "invalid-int",
	}

	// Valid key, return value.
	assert.Equal(t, ParseIntScreenParam(params, "key", int64(1)), int64(100))

	// Invalid key, return default value.
	assert.Equal(t, ParseIntScreenParam(params, "invaldKey", int64(1)),
		int64(1))

	// Valid key with invalid int value, return default value.
	assert.Equal(t, ParseIntScreenParam(params, "invalidIntKey", int64(1)),
		int64(1))
}
