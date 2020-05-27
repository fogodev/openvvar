package openvvar

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// fieldConfig holds informations about a struct field.
type fieldConfig struct {
	Name        string
	Short       string
	Key         string
	Description string
	Value       reflect.Value
	Default     reflect.Value
	Required    bool
}

var durationType = reflect.TypeOf(time.Duration(0))

// Set and String so fieldConfig complain with flag.Value
func (f *fieldConfig) Set(data string) error {
	return convert(data, f.Value)
}

func (f *fieldConfig) String() string {
	if f.Required && f.Default.IsZero() {
		return ""
	}
	return fmt.Sprintf("%v", f.Default)
}

func convert(data string, value reflect.Value) error {
	valueType := value.Type()

	// Duration is a special type because we need to reflect on an instance of it
	if valueType == durationType {
		d, err := time.ParseDuration(data)
		if err != nil {
			return &TypeConversionError{err}
		}
		value.SetInt(int64(d))
	} else {
		switch valueType.Kind() {
		case reflect.Bool:
			b, err := strconv.ParseBool(data)
			if err != nil {
				return &TypeConversionError{err}
			}
			value.SetBool(b)
		case reflect.Slice:
			// create a new temporary slice to override the actual Value if it's not empty
			splattedStrings := strings.Split(data, ",")
			newSlice := reflect.MakeSlice(value.Type(), 0, len(splattedStrings))
			for _, str := range splattedStrings {
				// create a new Value v based on the type of the slice
				currentValue := reflect.Indirect(reflect.New(valueType.Elem()))
				// call convert to set the current value of the slice to v
				if err := convert(str, currentValue); err != nil {
					return err // This one is an error of a recursive call
				}
				// append v to the temporary slice
				newSlice = reflect.Append(newSlice, currentValue)
			}
			// Set the newly created temporary slice to the target Value
			value.Set(newSlice)
		case reflect.String:
			value.SetString(data)
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			parsedInt, err := strconv.ParseInt(data, 10, valueType.Bits())
			if err != nil {
				return &TypeConversionError{err}
			}

			value.SetInt(parsedInt)
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			parsedUint, err := strconv.ParseUint(data, 10, valueType.Bits())
			if err != nil {
				return &TypeConversionError{err}
			}

			value.SetUint(parsedUint)
		case reflect.Float32, reflect.Float64:
			parsedFloat, err := strconv.ParseFloat(data, valueType.Bits())
			if err != nil {
				return &TypeConversionError{err}
			}
			value.SetFloat(parsedFloat)
		default:
			return &InvalidTypeForDefaultValuesError{valueType.Kind().String()}
		}
	}

	return nil
}
