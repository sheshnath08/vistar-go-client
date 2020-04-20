package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStringValue(t *testing.T) {
	params := map[string]interface{}{
		"key":    "test",
		"number": int64(20),
	}

	// Valid key with no override value, return value.
	v, err := ParseStringValue(params, "key", "default-value", "")
	assert.Nil(t, err)
	assert.Equal(t, v, "test")

	// Invalid key with no override value, return default value.
	v, err = ParseStringValue(params, "invaldKey", "def", "")
	assert.Nil(t, err)
	assert.Equal(t, v, "def")

	// Valid key with override value, return override value.
	v, err = ParseStringValue(params, "key", "def", "override-value")
	assert.Nil(t, err)
	assert.Equal(t, v, "override-value")

	// Invalid key with override value, return override value.
	v, err = ParseStringValue(params, "invalid-key", "def", "override-value")
	assert.Nil(t, err)
	assert.Equal(t, v, "override-value")

	// Valid  key with invalid number value, return error.
	v, err = ParseStringValue(params, "number", "default-value", "")
	assert.NotNil(t, err)
	assert.Equal(t, v, "")

}

func TestParseIntValue(t *testing.T) {
	params := map[string]interface{}{
		"key":                 float64(100),
		"invalid-number-type": false,
	}

	// Valid key with no override value, return value.
	v, err := ParseIntValue(params, "key", int64(200), int64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(100))

	// Invalid key with no override value, return default value.
	v, err = ParseIntValue(params, "invaldKey", int64(200), int64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(200))

	// Valid key with override value, return override value.
	v, err = ParseIntValue(params, "key", int64(200), int64(300))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(300))

	// Invalid key with override value, return override value.
	v, err = ParseIntValue(params, "invalidKey", int64(200), int64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(200))

	// Valid key with invalid number value, return error.
	v, err = ParseIntValue(params, "invalid-number-type", int64(200), int64(0))
	assert.NotNil(t, err)
	assert.Equal(t, v, int64(0))
}

func TestParseFloatValue(t *testing.T) {
	params := map[string]interface{}{
		"key":                 float64(100),
		"invalid-number-type": false,
	}

	// Valid key with no override value, return value.
	v, err := ParseFloatValue(params, "key", float64(200), float64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, float64(100))

	// Invalid key with no override value, return default value.
	v, err = ParseFloatValue(params, "invaldKey", float64(200), float64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, float64(200))

	// Valid key with override value, return override value.
	v, err = ParseFloatValue(params, "key", float64(200), float64(300))
	assert.Nil(t, err)
	assert.Equal(t, v, float64(300))

	// Invalid key with override value, return override value.
	v, err = ParseFloatValue(params, "invalidKey", float64(200), float64(0))
	assert.Nil(t, err)
	assert.Equal(t, v, float64(200))

	// Valid key with invalid number value, return error.
	v, err = ParseFloatValue(params, "invalid-number-type", float64(200),
		float64(0))
	assert.NotNil(t, err)
	assert.Equal(t, v, float64(0))
}

func TestParseBoolValue(t *testing.T) {
	params := map[string]interface{}{
		"trueValueKey":  true,
		"falseValueKey": false,
		"invalidBool":   "100",
	}

	// Valid key with no override value, return value.
	v, err := ParseBoolValue(params, "trueValueKey", false, false)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	v, err = ParseBoolValue(params, "falseValueKey", true, false)
	assert.Nil(t, err)
	assert.Equal(t, v, false)

	// Invalid key with no override value, return default value.
	v, err = ParseBoolValue(params, "invaldKey", true, false)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	// Valid key with override value, return override value.
	v, err = ParseBoolValue(params, "falseValueKey", false, true)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	// Invalid key with override value, return override value.
	v, err = ParseBoolValue(params, "falseValueKey", false, true)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	// valid key with invalid bool value, return error.
	v, err = ParseBoolValue(params, "invalidBool", true, false)
	assert.NotNil(t, err)
	assert.Equal(t, v, false)
}

func TestParseArrayValue(t *testing.T) {
	params := map[string]interface{}{
		"key":          []string{"test-1", "test-2"},
		"invalidArray": false,
	}

	defaultValue := []string{"default-1", "default-2"}
	overrideValue := []string{"override-1", "override-2"}

	// Valid key with no override value, return value.
	v, err := ParseArrayValue(params, "key", defaultValue, nil)
	assert.Nil(t, err)
	assert.Equal(t, v, []string{"test-1", "test-2"})

	// Invalid key with no override value, return default value.
	v, err = ParseArrayValue(params, "invaldKey", defaultValue, nil)
	assert.Nil(t, err)
	assert.Equal(t, v, []string{"default-1", "default-2"})

	// Valid key with override value, return override value.
	v, err = ParseArrayValue(params, "key", defaultValue, overrideValue)
	assert.Nil(t, err)
	assert.Equal(t, v, []string{"override-1", "override-2"})

	// Invalid key with override value, return override value.
	v, err = ParseArrayValue(params, "invalidKey", defaultValue, overrideValue)
	assert.Nil(t, err)
	assert.Equal(t, v, []string{"override-1", "override-2"})

	// Valid key with invalid array value, return error.
	v, err = ParseArrayValue(params, "invalidArray", defaultValue, nil)
	assert.NotNil(t, err)
	assert.Nil(t, v)
}

func TestParseBoolScreenParam(t *testing.T) {
	params := map[string]interface{}{
		"trueValueKey":  "true",
		"falseValueKey": "false",
		"key":           "invalid-bool",
	}

	// Valid key, return value.
	v, err := ParseBoolScreenParam(params, "trueValueKey", false)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	v, err = ParseBoolScreenParam(params, "falseValueKey", true)
	assert.Nil(t, err)
	assert.Equal(t, v, false)

	// Invalid key, return default value.
	v, err = ParseBoolScreenParam(params, "invaldKey", true)
	assert.Nil(t, err)
	assert.Equal(t, v, true)

	// Valid key with invalid bool value, return error.
	v, err = ParseBoolScreenParam(params, "key", false)
	assert.NotNil(t, err)
	assert.Equal(t, v, false)
}

func TestParseIntScreenParam(t *testing.T) {
	params := map[string]interface{}{
		"key":           "100",
		"invalidIntKey": "invalid-int",
	}

	// Valid key, return value.
	v, err := ParseIntScreenParam(params, "key", int64(1))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(100))

	// Invalid key, return default value.
	v, err = ParseIntScreenParam(params, "invaldKey", int64(1))
	assert.Nil(t, err)
	assert.Equal(t, v, int64(1))

	// Valid key with invalid int value, return error.
	v, err = ParseIntScreenParam(params, "invalidIntKey", int64(1))
	assert.NotNil(t, err)
	assert.Equal(t, v, int64(0))
}
