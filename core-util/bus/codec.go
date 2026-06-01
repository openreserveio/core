package bus

import (
	"context"
	"encoding/json"

	"github.com/openreserveio/core/core-util/otel"
)

func Encode(ctx context.Context, data interface{}) []byte {

	ctx = otel.StartSpan(ctx, "bus.Encode")
	defer otel.EndSpan(ctx)

	otel.AddEvent("Marshalling Data")
	bytes, err := json.Marshal(data)
	if err != nil {
		otel.AddError("Error marshalling data", err)
		return nil
	}

	return bytes

}

func Decode(ctx context.Context, data []byte, target interface{}) error {

	ctx = otel.StartSpan(ctx, "bus.Decode")
	defer otel.EndSpan(ctx)

	otel.AddEvent("Unmarshalling Data")
	err := json.Unmarshal(data, target)
	if err != nil {
		return err
	}

	return nil

}

func EncodeError(err error) []byte {

	errObj := make(map[string]string)
	errObj["error"] = err.Error()
	errJson, _ := json.Marshal(&errObj)

	return errJson

}
