package main

import (
	"context"

	"github.com/openreserveio/core/core-ledger-poster/application/coreledgerposter/cmd"
)

func main() {
	cmd.Execute(context.Background())
}
