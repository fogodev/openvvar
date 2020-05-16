package openvvar

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
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
func Load(receiverStruct interface{}, envFiles ...string) {

	if err := godotenv.Load(envFiles...); err != nil {
		// We're ignoring not found errors from default .env file
		if err.(*os.PathError).Path != ".env" {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	reflected := reflect.ValueOf(receiverStruct)

	if !reflected.IsValid() || reflected.Kind() != reflect.Ptr || reflected.Elem().Kind() != reflect.Struct {
		fmt.Fprintln(os.Stderr, "Provided config receiver must be a pointer to struct!")
		os.Exit(1)
	}

	fillData(parseStruct(reflected.Elem(), ""))
}

func parseStruct(receiverStruct reflect.Value, prefix string) *StructConfig {
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
				structConfig.Fields = append(structConfig.Fields, parseStruct(value, field.Name).Fields...)
				continue
			case reflect.Ptr:
				if valueType.Elem().Kind() == reflect.Struct && !value.IsNil() {
					structConfig.Fields = append(structConfig.Fields, parseStruct(value.Elem(), field.Name).Fields...)
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
				if idx := strings.Index(tag, ","); idx != -1 {
					fieldConfig.Key = tag[:idx]

					for _, opt := range strings.Split(tag[idx+1:], ",") {
						if opt == "required" {
							fieldConfig.Required = true
						} else if strings.HasPrefix(opt, shortString) {
							fieldConfig.Short = opt[len(shortString):]
						} else if strings.HasPrefix(opt, descriptionString) {
							fieldConfig.Description = opt[len(descriptionString):]
						} else if strings.HasPrefix(opt, defaultString) {
							fieldType := fieldConfig.Default.Kind()
							switch fieldType {

							case reflect.Bool:
								b, err := strconv.ParseBool(opt[len(defaultString):])
								if err != nil {
									fmt.Fprintf(os.Stderr, "Failed to parse boolean value %s on config %s", opt[len(defaultString):], fieldConfig.Key)
									os.Exit(1)
								}
								fieldConfig.Default.SetBool(b)

							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								i, err := strconv.ParseInt(opt[len(defaultString):], 10, 64)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Failed to parse integer value %s on config %s", opt[len(defaultString):], fieldConfig.Key)
									os.Exit(1)
								}
								fieldConfig.Default.SetInt(i)

							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
								u, err := strconv.ParseUint(opt[len(defaultString):], 10, 64)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Failed to parse unsigned integer value %s on config %s", opt[len(defaultString):], fieldConfig.Key)
									os.Exit(1)
								}
								fieldConfig.Default.SetUint(u)

							case reflect.Float32, reflect.Float64:
								f, err := strconv.ParseFloat(opt[len(defaultString):], 64)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Failed to parse floating point value %s on config %s", opt[len(defaultString):], fieldConfig.Key)
									os.Exit(1)
								}
								fieldConfig.Default.SetFloat(f)

							case reflect.String:
								fieldConfig.Default.SetString(opt[len(defaultString):])
							default:
								fmt.Fprintf(os.Stderr, "Default value not supported for type %s on field %s", fieldType, fieldConfig.Key)
								os.Exit(1)
							}
						}
					}
				}

				structConfig.Fields = append(structConfig.Fields, &fieldConfig)
			}
		}
	}

	return &structConfig
}

func fillData(structConfig *StructConfig) {

	LoadStruct(structConfig)

	for _, field := range structConfig.Fields {
		if field.Required && field.Value.IsZero() {
			fmt.Fprintf(os.Stderr, "Required key '%s' for field '%s' not found", field.Key, field.Name)
			os.Exit(1)
		}
	}
}
