package main

import (
	"context"

	"github.com/openreserveio/core/core-external-api-service/application/extapiservice/cmd"
)

func main() {
	cmd.Execute(context.Background())
}
