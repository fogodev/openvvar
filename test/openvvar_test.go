package test

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	openvvar "github.com/fogodev/openvvar/pkg"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	type nested struct {
		Int    int    `config:"int"`
		String string `config:"string"`
	}

	type testStruct struct {
		Bool            bool          `config:"bool"`
		Int             int           `config:"int"`
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
		Ptr             *string       `config:"ptr"`
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

	boolEnv := "true"

	intEnv := strconv.FormatInt(math.MaxInt64, 10)
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

	ptr := "ptr"
	stringFlag := "string"
	durationFlag := "10s"

	os.Setenv("BOOL", boolEnv)
	os.Setenv("INT", intEnv)
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

	os.Args = append(
		os.Args[:1],
		fmt.Sprintf("-float32=%s", float32Flag),
		fmt.Sprintf("-float64=%s", float64Flag),
		fmt.Sprintf("-ptr=%s", ptr),
		fmt.Sprintf("-string=%s", stringFlag),
		fmt.Sprintf("-duration=%s", durationFlag),
		fmt.Sprintf("-struct-int=%s", intEnv),
		fmt.Sprintf("-struct-string=%s", stringFlag),
		fmt.Sprintf("-structptrnotnil-int=%s", intEnv),
		fmt.Sprintf("-structptrnotnil-string=%s", stringFlag),
	)

	openvvar.Load(&s)

	require.EqualValues(t, testStruct{
		Bool:     true,
		Int:      math.MaxInt64,
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
		Ptr:      &ptr,
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
	if os.Getenv("TEST_LOAD_REQUIRED") == "1" {
		s := struct {
			Name string `config:"name,required"`
		}{}

		openvvar.Load(&s)
	} else {
		cmd := exec.Command(os.Args[0], "-test.run=TestLoadRequired")
		cmd.Env = append(os.Environ(), "TEST_LOAD_REQUIRED=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatal("Openvvar must exit on missing required field")
	}
}

func TestFlagPriority(t *testing.T) {

	type testStruct struct {
		Name string `config:"another_name,required"`
	}
	s := testStruct{}

	os.Args = append(os.Args[:1], fmt.Sprintf("-another_name=%s", "right"))
	os.Setenv("ANOTHER_NAME", "wrong")

	openvvar.Load(&s)

	require.EqualValues(t, testStruct{Name: "right"}, s)
}

func TestDefaultValue(t *testing.T) {

	s := struct {
		Name   string `config:"default_name,default=Xablau"`
		Number int    `config:"default_number,default=42"`
	}{}

	openvvar.Load(&s)

	require.EqualValues(t, "Xablau", s.Name)
}

func TestDefaultValueInvalidType(t *testing.T) {
	if os.Getenv("TEST_DEFAULT_VALUE_INVALID_TYPE") == "1" {
		s := struct {
			InvalidType map[string]string `config:"default_name,default=Xablau"`
		}{}

		openvvar.Load(&s)

	} else {
		cmd := exec.Command(os.Args[0], "-test.run=TestDefaultValueInvalidType")
		cmd.Env = append(os.Environ(), "TEST_DEFAULT_VALUE_INVALID_TYPE=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatal("Openvvar must exit on default value for invalid field type")
	}
}

func TestDotEnvFile(t *testing.T) {
	s := struct {
		Test string `config:"test,required"`
	}{}

	openvvar.Load(&s, ".env.test")

	require.EqualValues(t, s.Test, "test")
}

func TestDefaultValueParseFail(t *testing.T) {
	if os.Getenv("TEST_DEFAULT_VALUE_PARSE_FAIL") == "1" {
		s := struct {
			InvalidType bool `config:"default_name,default=Xablau"`
		}{}

		openvvar.Load(&s)
	} else {
		cmd := exec.Command(os.Args[0], "-test.run=TestDefaultValueParseFail")
		cmd.Env = append(os.Environ(), "TEST_DEFAULT_VALUE_PARSE_FAIL=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatal("Openvvar must exit on default value parse error")
	}
}

func TestDotEnvFileWithNestedField(t *testing.T) {

	type Nested struct {
		Name string `config:"name,required"`
	}

	s := struct {
		Nested Nested
	}{}

	openvvar.Load(&s, ".env.nested.test")

	require.EqualValues(t, s.Nested.Name, "test")
}
