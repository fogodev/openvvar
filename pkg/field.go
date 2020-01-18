package openvvar

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FieldConfig holds informations about a struct field.
type FieldConfig struct {
	Name        string
	Short       string
	Key         string
	Description string
	Value       reflect.Value
	Default     reflect.Value
	Required    bool
}

var durationType = reflect.TypeOf(time.Duration(0))

// Set converts data into f.Value.
func (f *FieldConfig) Set(data string) error {
	convert(data, f.Value)
	return nil // Returning nil to flag.Value interface be happy.
}

func (f *FieldConfig) String() string {
	return f.Key
}

func (f *FieldConfig) Get() interface{} {
	return f.Default.Interface()
}


func convert(data string, value reflect.Value) {
	t := value.Type()

	// Duration is a special type because we need to reflect on an instance of it
	if t == durationType {
		d, err := time.ParseDuration(data)
		if err != nil {
			panic(err)
		}
		value.SetInt(int64(d))
	} else {
		switch t.Kind() {
		case reflect.Bool:
			b, err := strconv.ParseBool(data)
			if err != nil {
				panic(err)
			}
			value.SetBool(b)
		case reflect.Slice:
			// create a new temporary slice to override the actual Value if it's not empty
			nv := reflect.MakeSlice(value.Type(), 0, 0)
			ss := strings.Split(data, ",")
			for _, s := range ss {
				// create a new Value v based on the type of the slice
				v := reflect.Indirect(reflect.New(t.Elem()))
				// call convert to set the current value of the slice to v
				convert(s, v)
				// append v to the temporary slice
				nv = reflect.Append(nv, v)
			}
			// Set the newly created temporary slice to the target Value
			value.Set(nv)

		case reflect.String:
			value.SetString(data)
		case reflect.Ptr:
			n := reflect.New(value.Type().Elem())
			value.Set(n)
			convert(data, n.Elem())
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			i, err := strconv.ParseInt(data, 10, t.Bits())
			if err != nil {
				panic(err)
			}

			value.SetInt(i)
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			i, err := strconv.ParseUint(data, 10, t.Bits())
			if err != nil {
				panic(err)
			}

			value.SetUint(i)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(data, t.Bits())
			if err != nil {
				panic(err)
			}
			value.SetFloat(f)
		default:
			panic(fmt.Sprintf("field type '%s' not supported", t.Kind()))
		}
	}
}