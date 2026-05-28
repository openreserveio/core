package main

import (
	"context"

	"github.com/openreserveio/core/core-ledger/application/coreledger/cmd"
)

func main() {
	ctx := context.Background()
	cmd.Execute(ctx)
}
