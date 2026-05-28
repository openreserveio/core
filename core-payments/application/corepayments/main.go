package main

import (
	"context"

	"github.com/openreserveio/core/core-payments/application/corepayments/cmd"
)

func main() {
	cmd.Execute(context.Background())
}
