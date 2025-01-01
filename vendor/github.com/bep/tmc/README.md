<h2 align="center">Codec for a Typed Map</h2>
<p align="center">Provides round-trip serialization of typed Go maps.<p>
<p align="center"><a href="https://godoc.org/github.com/bep/tmc"><img src="https://godoc.org/github.com/bep/tmc?status.svg" /></a>
<a href="https://goreportcard.com/report/github.com/bep/tmc"><img src="https://goreportcard.com/badge/github.com/bep/tmc" /></a>
<a href="https://codecov.io/gh/bep/tmc"><img src="https://codecov.io/gh/bep/tmc/branch/master/graph/badge.svg" /></a>
<a href="https://github.com/bep/tmc/actions"><img src="https://action-badges.now.sh/bep/tmc?workflow=test" /></a></p>


### How to Use

See the [GoDoc](https://godoc.org/github.com/bep/tmc) for some basic examples and how to configure custom codec, adapters etc.

### Why?

Text based serialization formats like JSON and YAML are convenient, but when used with Go maps, most type information gets lost in translation.

Listed below is a round-trip example in JSON (see https://play.golang.org/p/zxt-wi4Ljz3 for a runnable version):

```go
package main

import (
	"encoding/json"
	"log"
	"math/big"
	"time"

	"github.com/kr/pretty"
)

func main() {
	mi := map[string]interface{}{
		"vstring":   "Hello",
		"vint":      32,
		"vrat":      big.NewRat(1, 2),
		"vtime":     time.Now(),
		"vduration": 3 * time.Second,
		"vsliceint": []int{1, 3, 4},
		"nested": map[string]interface{}{
			"vint":      55,
			"vduration": 5 * time.Second,
		},
		"nested-typed-int": map[string]int{
			"vint": 42,
		},
		"nested-typed-duration": map[string]time.Duration{
			"v1": 5 * time.Second,
			"v2": 10 * time.Second,
		},
	}

	data, err := json.Marshal(mi)
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		log.Fatal(err)
	}

	pretty.Print(m)

}
```

This prints:

```go
map[string]interface {}{
    "vint":      float64(32),
    "vrat":      "1/2",
    "vtime":     "2009-11-10T23:00:00Z",
    "vduration": float64(3e+09),
    "vsliceint": []interface {}{
        float64(1),
        float64(3),
        float64(4),
    },
    "vstring": "Hello",
    "nested":  map[string]interface {}{
        "vduration": float64(5e+09),
        "vint":      float64(55),
    },
    "nested-typed-duration": map[string]interface {}{
        "v2": float64(1e+10),
        "v1": float64(5e+09),
    },
    "nested-typed-int": map[string]interface {}{
        "vint": float64(42),
    },
}
```

And that is very different from the origin:

* All numbers are now `float64`
* `time.Duration` is also `float64`
* `time.Now` and `*big.Rat` are strings
* Slices are `[]interface {}`, maps `map[string]interface {}`

So, for structs, you can work around some of the limitations above with custom `MarshalJSON`, `UnmarshalJSON`, `MarshalText` and `UnmarshalText`. 

For the commonly used flexible and schema-less`map[string]interface {}` this is, as I'm aware of, not an option.

Using this library, the above can be written to (see https://play.golang.org/p/PlDetQP5aWd for a runnable example):

```go
package main

import (
	"log"
	"math/big"
	"time"

	"github.com/bep/tmc"

	"github.com/kr/pretty"
)

func main() {
	mi := map[string]interface{}{
		"vstring":   "Hello",
		"vint":      32,
		"vrat":      big.NewRat(1, 2),
		"vtime":     time.Now(),
		"vduration": 3 * time.Second,
		"vsliceint": []int{1, 3, 4},
		"nested": map[string]interface{}{
			"vint":      55,
			"vduration": 5 * time.Second,
		},
		"nested-typed-int": map[string]int{
			"vint": 42,
		},
		"nested-typed-duration": map[string]time.Duration{
			"v1": 5 * time.Second,
			"v2": 10 * time.Second,
		},
	}

	c, err := tmc.New()
	if err != nil {
		log.Fatal(err)
	}

	data, err := c.Marshal(mi)
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]interface{})
	if err := c.Unmarshal(data, &m); err != nil {
		log.Fatal(err)
	}

	pretty.Print(m)

}
```

This prints:

```go
map[string]interface {}{
    "vduration":        time.Duration(3000000000),
    "vint":             int(32),
    "nested-typed-int": map[string]int{"vint":42},
    "vsliceint":        []int{1, 3, 4},
    "vstring":          "Hello",
    "vtime":            time.Time{
        wall: 0x0,
        ext:  63393490800,
        loc:  (*time.Location)(nil),
    },
    "nested": map[string]interface {}{
        "vduration": time.Duration(5000000000),
        "vint":      int(55),
    },
    "nested-typed-duration": map[string]time.Duration{"v1":5000000000, "v2":10000000000},
    "vrat":                  &big.Rat{
        a:  big.Int{
            neg: false,
            abs: {0x1},
        },
        b:  big.Int{
            neg: false,
            abs: {0x2},
        },
    },
}
```


### Performance

The implementation is easy to reason aobut (it uses reflection), but It's not particulary fast and probably not suited for _big data_. A simple benchmark with a roundtrip marshal/unmarshal is included. On my MacBook it shows:

```bash
BenchmarkCodec/JSON_regular-4         	   50000	     27523 ns/op	    6742 B/op	     171 allocs/op
BenchmarkCodec/JSON_typed-4           	   20000	     66644 ns/op	   16234 B/op	     411 allocs/op
```
