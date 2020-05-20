package openvvar

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

// StructConfig holds informations about each field of a struct S.
type StructConfig struct {
	Struct interface{}
	Fields []*FieldConfig
}

const shortString string = "short="
const descriptionString string = "description="
const defaultString string = "default="

// Load analyses all the Fields of the given struct for a "config" tag and queries flags and env vars
func Load(receiverStruct interface{}, envFiles ...string) error {

	if err := godotenv.Load(envFiles...); err != nil {
		// We're ignoring not found errors from default .env file
		var pathError *os.PathError
		if !errors.As(err, &pathError) || pathError.Path != ".env" {
			return &DotEnvNotFoundError{err}
		}
	}

	reflected := reflect.ValueOf(receiverStruct)

	if !reflected.IsValid() || reflected.Kind() != reflect.Ptr || reflected.Elem().Kind() != reflect.Struct {
		return &InvalidReceiver{}

	}

	if structConfig, err := parseStruct(reflected.Elem(), ""); err != nil {
		return err
	} else {
		return fillData(structConfig)
	}

}

func parseStruct(receiverStruct reflect.Value, prefix string) (*StructConfig, error) {
	var structConfig StructConfig

	receiverStructType := receiverStruct.Type()

	numFields := receiverStruct.NumField()
	for i := 0; i < numFields; i++ {
		field, value := receiverStructType.Field(i), receiverStruct.Field(i)
		valueType := value.Type()

		// Skipping current field if it is unexported
		if field.PkgPath == "" {
			// We're using "config" as out struct tag name
			tag := field.Tag.Get("config")

			// If current field is a struct or *struct, parse recursively using field name as prefix
			switch valueType.Kind() {
			case reflect.Struct:
				recursiveField, err := parseStruct(value, field.Name)
				if err != nil {
					return nil, err
				}

				structConfig.Fields = append(structConfig.Fields, recursiveField.Fields...)
				continue
			case reflect.Ptr:
				if valueType.Elem().Kind() == reflect.Struct && !value.IsNil() {
					recursiveField, err := parseStruct(value.Elem(), field.Name)
					if err != nil {
						return nil, err
					}
					structConfig.Fields = append(structConfig.Fields, recursiveField.Fields...)
					continue
				}
			}

			// Skipping fields with empty tags or no tags at all
			if tag != "" {
				if prefix != "" {
					tag = fmt.Sprintf("%s-%s", strings.ToLower(prefix), tag)
				}

				fieldConfig := FieldConfig{
					Name:  fmt.Sprintf("%s%s", prefix, field.Name),
					Key:   tag,
					Value: value,
				}

				// copying field content to a new value
				clone := reflect.Indirect(reflect.New(fieldConfig.Value.Type()))
				clone.Set(fieldConfig.Value)
				fieldConfig.Default = clone

				// Getting options for current field
				if idx := strings.Index(tag, ";"); idx != -1 {
					fieldConfig.Key = tag[:idx]

					for _, opt := range strings.Split(tag[idx+1:], ";") {
						if opt == "required" {
							fieldConfig.Required = true
						} else if strings.HasPrefix(opt, shortString) {
							fieldConfig.Short = opt[len(shortString):]
						} else if strings.HasPrefix(opt, descriptionString) {
							fieldConfig.Description = opt[len(descriptionString):]
						} else if strings.HasPrefix(opt, defaultString) {
							if err := convert(opt[len(defaultString):], fieldConfig.Default); err != nil {
								return nil, err
							}
						}
					}
				}

				structConfig.Fields = append(structConfig.Fields, &fieldConfig)
			}
		}
	}

	return &structConfig, nil
}

func fillData(structConfig *StructConfig) error {

	if err := loadStructData(structConfig); err != nil {
		return err
	}

	for _, field := range structConfig.Fields {
		if field.Required && field.Value.IsZero() {
			return &MissingRequiredFieldError{field.Key, field.Name}
		}
	}

	return nil
}
