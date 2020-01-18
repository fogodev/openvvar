package openvvar

import (
	"fmt"
	"reflect"
	"strings"
)

// StructConfig holds informations about each field of a struct S.
type StructConfig struct {
	S      interface{}
	Fields []*FieldConfig
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

	LoadStruct(structConfig)

	for _, field := range structConfig.Fields {
		if field.Required && isZero(field.Value) {
			panic(fmt.Sprintf("Required key '%s' for field '%s' not found", field.Key, field.Name))
		}
	}
}



func isZero(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	current := v.Interface()
	return reflect.DeepEqual(current, zero)
}
