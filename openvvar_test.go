package openvvar

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	type nested struct {
		Int    int    `config:"int"`
		String string `config:"string"`
	}

	type testStruct struct {
		Bool            bool          `config:"bool;short=b;description=Boolean for testing"`
		Int             int           `config:"int"`
		IntSlice        []int         `config:"int-slice"`
		Int8            int8          `config:"int8"`
		Int16           int16         `config:"int16"`
		Int32           int32         `config:"int32"`
		Int64           int64         `config:"int64"`
		Uint            uint          `config:"uint"`
		Uint8           uint8         `config:"uint8"`
		Uint16          uint16        `config:"uint16"`
		Uint32          uint32        `config:"uint32"`
		Uint64          uint64        `config:"uint64"`
		Float32         float32       `config:"float32"`
		Float64         float64       `config:"float64"`
		String          string        `config:"string"`
		Duration        time.Duration `config:"duration"`
		Struct          nested
		StructPtrNil    *nested
		StructPtrNotNil *nested
		Ignored         string
		unexported      int `config:"ignore"`
	}

	var s testStruct
	s.StructPtrNotNil = new(nested)

	boolShortFlag := "true"

	intEnv := strconv.FormatInt(math.MaxInt64, 10)
	intSliceEnv := strings.Join([]string{
		strconv.FormatInt(math.MaxInt64, 10),
		strconv.FormatInt(math.MaxInt64, 10),
	}, ",")
	int8Env := strconv.FormatInt(math.MaxInt8, 10)
	int16Env := strconv.FormatInt(math.MaxInt16, 10)
	int32Env := strconv.FormatInt(math.MaxInt32, 10)
	int64Env := strconv.FormatInt(math.MaxInt64, 10)

	uintEnv := strconv.FormatUint(math.MaxUint64, 10)
	uint8Env := strconv.FormatUint(math.MaxUint8, 10)
	uint16Env := strconv.FormatUint(math.MaxUint16, 10)
	uint32Env := strconv.FormatUint(math.MaxUint32, 10)
	uint64Env := strconv.FormatUint(math.MaxUint64, 10)

	float32Flag := strconv.FormatFloat(math.MaxFloat32, 'f', 6, 32)
	float64Flag := strconv.FormatFloat(math.MaxFloat64, 'f', 6, 64)

	stringFlag := "string"
	durationFlag := "10s"

	os.Setenv("INT", intEnv)
	os.Setenv("INT_SLICE", intSliceEnv)
	os.Setenv("INT8", int8Env)
	os.Setenv("INT16", int16Env)
	os.Setenv("INT32", int32Env)
	os.Setenv("INT64", int64Env)
	os.Setenv("UINT", uintEnv)
	os.Setenv("UINT8", uint8Env)
	os.Setenv("UINT16", uint16Env)
	os.Setenv("UINT32", uint32Env)
	os.Setenv("UINT64", uint64Env)
	os.Setenv("STRUCTPTRNOTNIL_INT", intEnv)
	os.Setenv("STRUCTPTRNOTNIL_STRING", stringFlag)

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	os.Args = append(
		os.Args[:1],
		fmt.Sprintf("-b=%s", boolShortFlag),
		fmt.Sprintf("-float32=%s", float32Flag),
		fmt.Sprintf("-float64=%s", float64Flag),
		fmt.Sprintf("-string=%s", stringFlag),
		fmt.Sprintf("-duration=%s", durationFlag),
		fmt.Sprintf("-struct-int=%s", intEnv),
		fmt.Sprintf("-struct-string=%s", stringFlag),
		fmt.Sprintf("-structptrnotnil-int=%s", intEnv),
		fmt.Sprintf("-structptrnotnil-string=%s", stringFlag),
	)

	assert.Nil(t, Load(&s))

	assert.Equal(t, testStruct{
		Bool:     true,
		Int:      math.MaxInt64,
		IntSlice: []int{math.MaxInt64, math.MaxInt64},
		Int8:     math.MaxInt8,
		Int16:    math.MaxInt16,
		Int32:    math.MaxInt32,
		Int64:    math.MaxInt64,
		Uint:     math.MaxUint64,
		Uint8:    math.MaxUint8,
		Uint16:   math.MaxUint16,
		Uint32:   math.MaxUint32,
		Uint64:   math.MaxUint64,
		Float32:  math.MaxFloat32,
		Float64:  math.MaxFloat64,
		String:   "string",
		Duration: 10 * time.Second,
		Struct: nested{
			Int:    math.MaxInt64,
			String: "string",
		},
		StructPtrNotNil: &nested{
			Int:    math.MaxInt64,
			String: "string",
		},
	}, s)

	s.unexported = 1 // This line exist just to satisfy golangci-lint
}

func TestLoadRequired(t *testing.T) {

	s := struct {
		Name string `config:"name;required"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(t, errors.Is(Load(&s), &MissingRequiredFieldError{
		Key:   "name",
		Field: "Name",
	}), "Openvvar must throw error on missing required field")

}

func TestDefaultValue(t *testing.T) {

	type testStructDefaultValues struct {
		Bool     bool          `config:"default-bool;default=true"`
		String   string        `config:"default-string;default=Xablau"`
		Int      int           `config:"default-int;default=42"`
		IntSlice []int         `config:"default-int-slice;default=123,456"`
		Duration time.Duration `config:"default-duration;default=5s"`
		Uint     uint          `config:"default-uint;default=999"`
		Float    float64       `config:"default-float;default=42.999"`
	}

	s := testStructDefaultValues{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.Nil(t, Load(&s))

	assert.Equal(t, testStructDefaultValues{
		Bool:     true,
		String:   "Xablau",
		Int:      42,
		IntSlice: []int{123, 456},
		Duration: 5 * time.Second,
		Uint:     999,
		Float:    42.999,
	}, s)
}

func TestDefaultValueInvalidType(t *testing.T) {

	s := struct {
		InvalidType map[string]string `config:"default_name;default=Xablau"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(&s), &InvalidTypeForDefaultValuesError{Type: "map"}),
		"Openvvar must throw an InvalidTypeForDefaultValuesError error on default value for invalid field type",
	)

}

func TestDefaultValueParseFail(t *testing.T) {

	s := struct {
		InvalidBool     bool          `config:"invalid-bool;default=Xablau"`
		InvalidDuration time.Duration `config:"invalid-duration;default=Xablau"`
		InvalidInt      int           `config:"invalid-int;default=Xablau"`
		InvalidUint     uint          `config:"invalid-uint;default=Xablau"`
		InvalidFloat    float64       `config:"invalid-float;default=Xablau"`
		InvalidIntSlice []int         `config:"invalid-int-slice;default=Xablau,Xaplay"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(&s), &TypeConversionError{}),
		"Openvvar must throw an error for invalid default values",
	)

}

func TestDefaultValueParsePtrFail(t *testing.T) {

	type InvalidNested struct {
		Bool bool `config:"bool;default=Xablau"`
	}

	s := struct {
		InvalidPtr *InvalidNested
	}{InvalidPtr: new(InvalidNested)}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(&s, "test_samples/.env.invalid_types"), &TypeConversionError{}),
		"Openvvar must throw an error for invalid dot env values",
	)
}

func TestEnvValueParseFail(t *testing.T) {
	s := struct {
		InvalidBool     bool          `config:"invalid-bool"`
		InvalidDuration time.Duration `config:"invalid-duration"`
		InvalidInt      int           `config:"invalid-int"`
		InvalidUint     uint          `config:"invalid-uint"`
		InvalidFloat    float64       `config:"invalid-float"`
		InvalidIntSlice []int         `config:"invalid-int-slice"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	os.Setenv("INVALID_BOOL", "Xablau")
	os.Setenv("INVALID_DURATION", "Xablau")
	os.Setenv("INVALID_INT", "Xablau")
	os.Setenv("INVALID_UINT", "Xablau")
	os.Setenv("INVALID_FLOAT", "Xablau")
	os.Setenv("INVALID_INT_SLICE", "Xablau")

	assert.True(
		t,
		errors.Is(Load(&s), &FlagCollectionError{}),
		"Openvvar must throw an error for invalid default values",
	)
}

func TestDotEnvFile(t *testing.T) {
	s := struct {
		Test string `config:"test;required"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.Nil(t, Load(&s, "test_samples/.env.test"))

	assert.Equal(t, s.Test, "test")
}

func TestDotEnvValueParseStructFail(t *testing.T) {

	type InvalidStruct struct {
		Invalid bool `config:"bool;default=Xablau"`
	}

	s := struct {
		InvalidStruct InvalidStruct
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(&s, "test_samples/.env.invalid_types"), &TypeConversionError{}),
		"Openvvar must throw an error for invalid dot env values",
	)
}

func TestDotEnvFileWithNestedField(t *testing.T) {

	type Inner struct {
		Name string `config:"name;required"`
	}

	s := struct {
		Inner Inner
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.Nil(t, Load(&s, "test_samples/.env.nested.test"))

	assert.Equal(t, "test", s.Inner.Name)
}

func TestFlagPriority(t *testing.T) {

	type testStruct struct {
		Name string `config:"another_name;required"`
	}
	s := testStruct{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	os.Args = append(os.Args[:1], fmt.Sprintf("-another_name=%s", "right"))
	os.Setenv("ANOTHER_NAME", "wrong")

	assert.Nil(t, Load(&s))

	assert.Equal(t, testStruct{Name: "right"}, s)
}

func TestFlagsValueParseFail(t *testing.T) {

	s := struct {
		InvalidBool     bool          `config:"invalid-bool"`
		InvalidDuration time.Duration `config:"invalid-duration"`
		InvalidInt      int           `config:"invalid-int"`
		InvalidUint     uint          `config:"invalid-uint"`
		InvalidFloat    float64       `config:"invalid-float"`
		InvalidIntSlice []int         `config:"invalid-int-slice"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	os.Args = append(
		os.Args,
		"-invalid-bool=Xablau",
		"-invalid-duration=Xablau",
		"-invalid-int=Xablau",
		"-invalid-uint=Xablau",
		"-invalid-float=Xablau",
		"-invalid-int-slice=Xablau",
	)

	assert.True(
		t,
		errors.Is(Load(&s), &FlagParseError{}),
		"Openvvar must throw an error for invalid default values",
	)
}

func TestNotFoundEnvFile(t *testing.T) {

	s := struct {
		NotFound string `config:"default_name;default=Xablau"`
	}{}

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(&s, ".env.not_found"), &DotEnvNotFoundError{}),
		"Openvvar must exit on not found env file",
	)

}

func TestInvalidReceiver(t *testing.T) {

	// Cleaning args so a test args can't interfere on another test
	os.Args = os.Args[:1]

	assert.True(
		t,
		errors.Is(Load(42), &InvalidReceiver{}),
		"Openvvar must throw an error on invalid receiver",
	)
}

func TestErrors(t *testing.T) {
	notFound := errors.New("not found")
	assert.Equal(t, (&DotEnvNotFoundError{notFound}).Error(), "not found")
	assert.True(t, errors.Is((&DotEnvNotFoundError{notFound}).Unwrap(), notFound))
	assert.False(t, errors.Is(&DotEnvNotFoundError{notFound}, errors.New("other error")))

	errorsSet := map[error]bool{
		errors.New("error 1"): true,
		errors.New("error 2"): true,
		errors.New("error 3"): true,
	}

	assert.Equal(t, (&FlagCollectionError{map[error]bool{errors.New("single error"): true}}).Error(), "single error")
	assert.False(t, errors.Is(&FlagCollectionError{}, notFound))
	assert.False(t, errors.Is(&FlagCollectionError{errorsSet}, &FlagCollectionError{map[error]bool{}}))
	assert.True(t, errors.Is(&FlagCollectionError{errorsSet}, &FlagCollectionError{errorsSet}))

	parseError := errors.New("parse error")
	assert.Equal(t, (&FlagParseError{parseError}).Error(), "parse error")
	assert.False(t, errors.Is(&FlagParseError{}, notFound))

	conversionError := errors.New("conversion error")
	assert.Equal(t, (&TypeConversionError{conversionError}).Error(), "conversion error")
	assert.Equal(t, (&TypeConversionError{conversionError}).Unwrap(), conversionError)
	assert.False(t, errors.Is(&TypeConversionError{}, conversionError))

	assert.Equal(t, (&MissingRequiredFieldError{"a", "b"}).Error(), "required key 'a' for field 'b' not found")
	assert.False(t, errors.Is(&MissingRequiredFieldError{}, conversionError))

	assert.Equal(t, (&InvalidTypeForDefaultValuesError{"a"}).Error(), "field type 'a' not supported")
	assert.False(t, errors.Is(&InvalidTypeForDefaultValuesError{}, conversionError))

	assert.Equal(t, (&InvalidReceiver{}).Error(), "provided config receiver must be a pointer to struct")
	assert.False(t, errors.Is(&InvalidReceiver{}, conversionError))
}
