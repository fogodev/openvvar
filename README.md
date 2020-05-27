[![github](https://github.com/fogodev/openvvar/workflows/Unit%20Tests/badge.svg)](https://github.com/fogodev/openvvar)
[![codecov](https://codecov.io/gh/fogodev/openvvar/branch/master/graph/badge.svg)](https://codecov.io/gh/fogodev/openvvar)
# openvvar - Opinionated Environment Variables

Package openvvar provides an easy way to manage flags and environment variables at same time.
Making use of struct tags to structure your configurations, providing neat features like nested structs
for correlated configurations, required fields, default values for all the "primitive" types, like ints, uints,
strings, booleans, floats, time.Duration and slices for any of those types.

### Usage

```go
package main

import (
    "fmt"
    "time"
    "os"
    "github.com/fogodev/openvvar"
)

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

func main() {
    configs := Config{}
    if err := openvvar.Load(&configs); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    /*
        ...
    */
}
```

Nested fields have their parent field name concatenated to its own name

```shell script
$ DATABASE_USER=root # For environment variables

$ ./your_program -database-password=1234 # for flags
```

To load configurations, just instantiate an object from your struct and pass its pointer to Load function,
checking for errors afterward, just like in the example above, or you can pass one or more paths to dot env files

For more examples check unit tests file
	