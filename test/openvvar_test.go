package test

import (
	"fmt"
	openvvar "github.com/fogodev/openvvar/pkg"
	"math"

	"flag"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type store map[string]string

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
	ptrFlag := fmt.Sprintf("%p", &ptr)
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


	flag.Set("float32", float32Flag)
	flag.Set("float64", float64Flag)
	flag.Set("ptr", ptrFlag)
	flag.Set("string", stringFlag)
	flag.Set("duration", durationFlag)

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

}