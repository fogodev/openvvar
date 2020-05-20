package openvvar

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// loadStructData takes a struct config, define flags based on it and parse the command line args.
func loadStructData(config *StructConfig) error {

	commandLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	for _, field := range config.Fields {
		field.Value.Set(field.Default)
		commandLine.Var(field, field.Key, field.Description)
		if field.Short != "" {
			commandLine.Var(field, field.Short, shortDesc(field.Description))
		}
	}

	var allErrors []error
	commandLine.VisitAll(func(f *flag.Flag) {
		if envVar, found := getEnvVar(f.Name); found {
			if err := commandLine.Set(f.Name, envVar); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	})

	if err := commandLine.Parse(os.Args[1:]); err != nil {
		return &FlagParseError{err}
	}

	if len(allErrors) > 0 {
		errorsSet := make(map[error]bool, len(allErrors))
		for _, err := range allErrors {
			errorsSet[err] = true
		}

		return &FlagCollectionError{Errors: errorsSet}
	}

	return nil
}

func shortDesc(description string) string {
	return fmt.Sprintf("%s (short)", description)
}

func getEnvVar(flagName string) (string, bool) {
	if val := os.Getenv(strings.Replace(strings.ToUpper(flagName), "-", "_", -1)); val != "" {
		return val, true
	}
	return "", false
}
