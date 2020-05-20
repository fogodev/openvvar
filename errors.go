package openvvar

import (
	"errors"
	"fmt"
	"strings"
)

type DotEnvNotFoundError struct {
	Err error
}

func (e *DotEnvNotFoundError) Error() string {
	return e.Err.Error()
}

func (e *DotEnvNotFoundError) Unwrap() error {
	return e.Err
}

func (e *DotEnvNotFoundError) Is(target error) bool {
	tar, ok := target.(*DotEnvNotFoundError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil
}

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

type FlagParseError struct {
	Err error
}

func (e *FlagParseError) Error() string {
	return e.Err.Error()
}

func (e *FlagParseError) Is(target error) bool {
	tar, ok := target.(*FlagParseError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil

}

type TypeConversionError struct {
	Err error
}

func (e *TypeConversionError) Error() string {
	return e.Err.Error()
}

func (e *TypeConversionError) Unwrap() error {
	return e.Err
}

func (e *TypeConversionError) Is(target error) bool {
	tar, ok := target.(*TypeConversionError)
	if !ok {
		return false
	}

	return errors.Is(e.Err, tar.Err) || tar.Err == nil
}

type MissingRequiredFieldError struct {
	Key   string
	Field string
}

func (e *MissingRequiredFieldError) Error() string {
	return fmt.Sprintf("required key '%s' for field '%s' not found", e.Key, e.Field)
}

func (e *MissingRequiredFieldError) Is(target error) bool {
	tar, ok := target.(*MissingRequiredFieldError)
	if !ok {
		return false
	}

	return (e.Field == tar.Field || tar.Field == "") && (e.Key == tar.Key || tar.Key == "")
}

type InvalidTypeForDefaultValuesError struct {
	Type string
}

func (e *InvalidTypeForDefaultValuesError) Error() string {
	return fmt.Sprintf("field type '%s' not supported", e.Type)
}

func (e *InvalidTypeForDefaultValuesError) Is(target error) bool {
	tar, ok := target.(*InvalidTypeForDefaultValuesError)
	if !ok {
		return false
	}

	return e.Type == tar.Type || tar.Type == ""
}

type InvalidReceiver struct{}

func (e *InvalidReceiver) Error() string {
	return "provided config receiver must be a pointer to struct"
}

func (e *InvalidReceiver) Is(target error) bool {
	_, ok := target.(*InvalidReceiver)
	return ok
}
