package openvvar

import (
	"flag"
	"fmt"
	"reflect"
	"time"
)

// LoadStruct takes a struct config, define flags based on it and parse the command line args.
func LoadStruct(cfg *StructConfig) {
	for _, loopingField := range cfg.Fields {
		field := loopingField
		// Display all the flags and their default values but override the field only if the user has explicitely
		// set the flag.
		k := field.Value.Kind()
		switch {
		case field.Value.Type().String() == "time.Duration":
			var val time.Duration
			flag.DurationVar(&val, field.Key, time.Duration(field.Default.Int()), field.Description)
			if field.Short != "" {
				flag.DurationVar(&val, field.Short, time.Duration(field.Default.Int()), shortDesc(field.Description))
			}
			// this function must be executed after the flag.Parse call.
			defer func() {
				// if the user has set the flag, save the value in the field.
				if isFlagSet(field) {
					field.Value.SetInt(int64(val))
				}
			}()
		case k == reflect.Bool:
			var val bool
			flag.BoolVar(&val, field.Key, field.Default.Bool(), field.Description)
			if field.Short != "" {
				flag.BoolVar(&val, field.Short, field.Default.Bool(), shortDesc(field.Description))
			}
			defer func() {
				if isFlagSet(field) {
					field.Value.SetBool(val)
				}
			}()
		case k >= reflect.Int && k <= reflect.Int64:
			var val int
			flag.IntVar(&val, field.Key, int(field.Default.Int()), field.Description)
			if field.Short != "" {
				flag.IntVar(&val, field.Short, int(field.Default.Int()), shortDesc(field.Description))
			}
			defer func() {
				if isFlagSet(field) {
					field.Value.SetInt(int64(val))
				}
			}()
		case k >= reflect.Uint && k <= reflect.Uint64:
			var val uint64
			flag.Uint64Var(&val, field.Key, field.Default.Uint(), field.Description)
			if field.Short != "" {
				flag.Uint64Var(&val, field.Short, field.Default.Uint(), shortDesc(field.Description))
			}
			defer func() {
				if isFlagSet(field) {
					field.Value.SetUint(val)
				}
			}()
		case k >= reflect.Float32 && k <= reflect.Float64:
			var val float64
			flag.Float64Var(&val, field.Key, field.Default.Float(), field.Description)
			if field.Short != "" {
				flag.Float64Var(&val, field.Short, field.Default.Float(), shortDesc(field.Description))
			}
			defer func() {
				if isFlagSet(field) {
					field.Value.SetFloat(val)
				}
			}()
		case k == reflect.String:
			var val string
			flag.StringVar(&val, field.Key, field.Default.String(), field.Description)
			if field.Short != "" {
				flag.StringVar(&val, field.Short, field.Default.String(), shortDesc(field.Description))
			}
			defer func() {
				if isFlagSet(field) {
					field.Value.SetString(val)
				}
			}()
		default:
			flag.Var(field, field.Key, field.Description)
		}
	}

	flagSet := map[string]bool{}
	flag.CommandLine.Visit(func(f *flag.Flag) {
		flagSet[f.Name] = true
	})

	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if _, defined := flagSet[f.Name]; !defined {
			if envVar, notFound := Get(f.Name); notFound == nil {
				if err := flag.CommandLine.Set(f.Name, envVar); err != nil {
					panic(err)
				}
			}
		}
	})

	flag.Parse()
}

func shortDesc(description string) string {
	return fmt.Sprintf("%s (short)", description)
}

func isFlagSet(config *FieldConfig) bool {
	flagSet := make(map[*FieldConfig]bool)
	flag.Visit(func(f *flag.Flag) { flagSet[config] = true })

	_, ok := flagSet[config]
	return ok
}
