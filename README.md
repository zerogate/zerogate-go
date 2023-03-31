# zerogate-go
[![Go Reference](https://pkg.go.dev/badge/github.com/zerogate/zerogate-go.svg)](https://pkg.go.dev/github.com/zerogate/zerogate-go)
![Test](https://github.com/zerogate/zerogate-go/workflows/Test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zerogate/zerogate-go?style=flat-square)](https://goreportcard.com/report/github.com/zerogate/zerogate-go)

A Go library for interacting with ZeroGate API.

## Getting Started

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zerogate/zerogate-go"
)

func main() {
	// Construct a new Client object using API key & API secret
	client, err := zerogate.New(os.Getenv("ZEROGATE_API_KEY"), os.Getenv("ZEROGATE_API_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	
	// Fetch all tenants
	tenants, total, err := client.Tenant.List(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	// Print tenant details
	fmt.Println(total)
	fmt.Println(tenants)
}
```

Also refer to the
[API documentation](https://pkg.go.dev/github.com/zerogate/zerogate-go) for
how to use this package in-depth.

## License

Mozilla Public License Version 2.0 licensed. See the [LICENSE](LICENSE) file for details.