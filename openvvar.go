/*
Package openvvar provides an easy way to manage flags and environment variables at same time.
Making use of struct tags to structure your configurations, providing neat features like nested structs
for correlated configurations, required fields, default values for all the "primitive" types, like ints, uints,
strings, booleans, floats, time.Duration and slices for any of those types.

	type DatabaseConfig struct {
		Name     string `config:"name;default=postgresql"`
		Host     string `config:"host;default=localhost"`
		Port     int    `config:"port;default=5432"`
		User     string `config:"user;required"`
		Password string `config:"password;required"`
	}

	type Config struct {
		Database          DatabaseConfig
		Debug             bool          `config:"debug;default=false;description=Set this config to true for debug log"`
		AcceptedHeroNames []string      `config:"hero-names;default=Deadpool,Iron Man,Dr. Strange,Rocket Raccon"`
		UniversalAnswer   uint8         `config:"universal-answer;default=42;short=u;description=THE ANSWER TO LIFE, THE UNIVERSE AND EVERYTHING"`
		SomeRandomFloat   float64       `config:"random-float;default=149714.1241"`
		OneSecond         time.Duration `config:"second;default=1s"`
	}

Nested fields have their parent field name concatenated to its own name

	$ DATABASE_USER=root # For environment variables

	$ ./your_program -database-password=1234 # for flags

To load configurations, just instantiate an object from your struct and pass its pointer to Load function,
checking for errors afterward:

	configs := Config{}
	if err := openvvar.Load(&configs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

*/
package openvvar

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

// structConfig holds information about each field of a struct S.
type structConfig struct {
	Struct interface{}
	Fields []*fieldConfig
}

const shortString string = "short="
const descriptionString string = "description="
const defaultString string = "default="
const optionsString string = "options="

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
		return &InvalidReceiverError{}

	}

	structConfig, err := parseStruct(reflected.Elem(), "")
	if err != nil {
		return err
	}

	return fillData(structConfig)

}

func parseStruct(receiverStruct reflect.Value, prefix string) (*structConfig, error) {
	var structConfig structConfig

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

				fieldConfig := fieldConfig{
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
						} else if strings.HasPrefix(opt, optionsString) {
							fieldConfig.Options = make(map[string]bool)
							for _, option := range strings.Split(opt[len(optionsString):], ",") {
								fieldConfig.Options[strings.TrimSpace(option)] = true
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

func fillData(structConfig *structConfig) error {

	if err := loadStructData(structConfig); err != nil {
		return err
	}

	for _, field := range structConfig.Fields {
		if field.Required && field.Value.IsZero() {
			return &MissingRequiredFieldError{field.Key, field.Name}
		}

		if field.Options != nil {
			if _, ok := field.Options[field.Value.String()]; !ok {
				return &ValueNotAValidOptionError{
					Value:   field.Value.String(),
					Options: field.Options,
				}
			}
		}
	}

	return nil
}
