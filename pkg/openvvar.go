package openvvar

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// StructConfig holds informations about each field of a struct S.
type StructConfig struct {
	S      interface{}
	Fields []*FieldConfig
}

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

// Set converts data into f.Value.
func (f *FieldConfig) Set(data string) error {
	convert(data, f.Value)
	return nil // Returning nil to flag.Value interface be happy.
}

// Load analyses all the Fields of the given struct for a "config" tag and queries flags and env vars
func Load(receiverStruct interface{}) {

	reflected := reflect.ValueOf(receiverStruct)

	if !reflected.IsValid() || reflected.Kind() != reflect.Ptr || reflected.Elem().Kind() != reflect.Struct {
		panic("Provided config receiver must be a pointer to struct!")
	}

	reflected = reflected.Elem()

	s := parseStruct(reflected, "")
	resolve(s)
}

func parseStruct(receiverStruct reflect.Value, prefix string) *StructConfig {
	var structConfig StructConfig

	t := receiverStruct.Type()

	numFields := receiverStruct.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		value := receiverStruct.Field(i)
		typ := value.Type()

		// skip if field is unexported
		if field.PkgPath == "" {
			tag := field.Tag.Get("config")

			// if struct or *struct, parse recursively
			switch typ.Kind() {
			case reflect.Struct:
				structConfig.Fields = append(structConfig.Fields, parseStruct(value, field.Name).Fields...)
				continue
			case reflect.Ptr:
				if typ.Elem().Kind() == reflect.Struct && !value.IsNil() {
					structConfig.Fields = append(structConfig.Fields, parseStruct(value.Elem(), field.Name).Fields...)
					continue
				}
			}

			// empty tag or no tag, skip the field
			if tag != "" {
				var key string
				if prefix != "" {
					key = strings.Join([]string{strings.ToLower(prefix), tag}, "-")
				} else {
					key = tag
				}

				fieldConfig := FieldConfig{
					Name:  fmt.Sprintf("%s%s", prefix, field.Name),
					Key:   key,
					Value: value,
				}

				// copying field content to a new value
				clone := reflect.Indirect(reflect.New(fieldConfig.Value.Type()))
				clone.Set(fieldConfig.Value)
				fieldConfig.Default = clone

				if idx := strings.Index(tag, ","); idx != -1 {
					fieldConfig.Key = tag[:idx]
					opts := strings.Split(tag[idx+1:], ",")

					for _, opt := range opts {
						if opt == "required" {
							fieldConfig.Required = true
						} else if strings.HasPrefix(opt, "short=") {
							fieldConfig.Short = opt[len("short="):]
						} else if strings.HasPrefix(opt, "description=") {
							fieldConfig.Description = opt[len("description="):]
						}
					}
				}

				structConfig.Fields = append(structConfig.Fields, &fieldConfig)
			}
		}
	}

	return &structConfig
}

func resolve(structConfig *StructConfig) {

	foundFields := make(map[*FieldConfig]bool)

	LoadStruct(structConfig)

	if len(foundFields) != len(structConfig.Fields) {
		for _, f := range structConfig.Fields {
			if _, ok := foundFields[f]; !ok {
				raw, err := Get(f.Key)
				if err == nil {
					convert(string(raw), f.Value)
					foundFields[f] = true
				}
			}
		}
	}

	for _, field := range structConfig.Fields {
		if field.Required && isZero(field.Value) {
			panic(fmt.Sprintf("Required key '%s' for field '%s' not found", field.Key, field.Name))
		}
	}
}



var durationType = reflect.TypeOf(time.Duration(0))

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

func isZero(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	current := v.Interface()
	return reflect.DeepEqual(current, zero)
}
