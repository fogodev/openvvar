package openvvar

import (
	"errors"
	"fmt"
	"strings"
)

// DotEnvNotFoundError for when the file is not found
type DotEnvNotFoundError struct {
	Err error
}

func (e *DotEnvNotFoundError) Error() string {
	return e.Err.Error()
}

func (e *DotEnvNotFoundError) Unwrap() error {
	return e.Err
}

// Is method to comply with new errors functions
func (e *DotEnvNotFoundError) Is(target error) bool {
	tar, ok := target.(*DotEnvNotFoundError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil
}

// FlagCollectionError when we receive errors from flag library
type FlagCollectionError struct {
	Errors map[error]bool // Using this map as a hash set
}

func (e *FlagCollectionError) Error() string {
	keys := make([]string, 0, len(e.Errors))
	for key := range e.Errors {
		keys = append(keys, key.Error())
	}

	return strings.Join(keys, ": ")
}

// Is method to comply with new errors functions
func (e *FlagCollectionError) Is(target error) bool {
	tar, ok := target.(*FlagCollectionError)
	if !ok {
		return false
	}
	if tar.Errors == nil {
		return true
	}

	for err, v := range e.Errors {
		if tar.Errors[err] != v {
			// If one of my errors doesn't exist on target, we're different errors
			return false
		}
	}

	return true
}

// FlagParseError is for when we fail to parse a specific flag
type FlagParseError struct {
	Err error
}

func (e *FlagParseError) Error() string {
	return e.Err.Error()
}

// Is method to comply with new errors functions
func (e *FlagParseError) Is(target error) bool {
	tar, ok := target.(*FlagParseError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil

}

// TypeConversionError occurs on string parsing for some types
type TypeConversionError struct {
	Err error
}

func (e *TypeConversionError) Error() string {
	return e.Err.Error()
}

func (e *TypeConversionError) Unwrap() error {
	return e.Err
}

// Is method to comply with new errors functions
func (e *TypeConversionError) Is(target error) bool {
	tar, ok := target.(*TypeConversionError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil
}

// MissingRequiredFieldError when user forgets to fill a required config
type MissingRequiredFieldError struct {
	Key   string
	Field string
}

func (e *MissingRequiredFieldError) Error() string {
	return fmt.Sprintf("required key '%s' for field '%s' not found", e.Key, e.Field)
}

// Is method to comply with new errors functions
func (e *MissingRequiredFieldError) Is(target error) bool {
	tar, ok := target.(*MissingRequiredFieldError)
	if !ok {
		return false
	}

	return (e.Field == tar.Field || tar.Field == "") && (e.Key == tar.Key || tar.Key == "")
}

// InvalidTypeForDefaultValuesError when developer puts a bogus default value for some type
type InvalidTypeForDefaultValuesError struct {
	Type string
}

func (e *InvalidTypeForDefaultValuesError) Error() string {
	return fmt.Sprintf("field type '%s' not supported", e.Type)
}

// Is method to comply with new errors functions
func (e *InvalidTypeForDefaultValuesError) Is(target error) bool {
	tar, ok := target.(*InvalidTypeForDefaultValuesError)
	if !ok {
		return false
	}

	return e.Type == tar.Type || tar.Type == ""
}

// InvalidReceiverError when developer pass something that isn't a point to struct to receive configs
type InvalidReceiverError struct{}

func (e *InvalidReceiverError) Error() string {
	return "provided config receiver must be a pointer to struct"
}

// Is method to comply with new errors functions
func (e *InvalidReceiverError) Is(target error) bool {
	_, ok := target.(*InvalidReceiverError)
	return ok
}

// ValueNotAValidOptionError for when received valued is not listed as a valid option
type ValueNotAValidOptionError struct {
	Value   string
	Options map[string]bool
}

func (e *ValueNotAValidOptionError) Error() string {
	options := make([]string, 0, len(e.Options))
	for option := range e.Options {
		options = append(options, option)
	}
	return fmt.Sprintf("received value \"%s\" is not a valid option from %v", e.Value, options)
}

// Is method to comply with new errors functions
func (e *ValueNotAValidOptionError) Is(target error) bool {
	tar, ok := target.(*ValueNotAValidOptionError)
	if !ok {
		return false
	}
	if tar.Options == nil && tar.Value == "" {
		return true
	}

	for option, v := range e.Options {
		if tar.Options[option] != v {
			// If one of my options doesn't exist on target, we're different errors
			return false
		}
	}

	return true
}
