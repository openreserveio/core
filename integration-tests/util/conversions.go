package util

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func ConvertPBAnyToMap(from map[string]*anypb.Any) map[string]interface{} {

	var converted = make(map[string]interface{})
	for key, value := range from {
		// convert value from anypb.Any to interface{}
		convertedValue, err := anypb.UnmarshalNew(value, proto.UnmarshalOptions{})
		if err != nil {
			log.Errorf("Unable to Unmashal anypb: %v", err)
			return nil
		}

		converted[key] = convertedValue
	}

	return converted

}

func ConvertMapToPBAny(from map[string]interface{}) map[string]*anypb.Any {

	var converted = make(map[string]*anypb.Any)
	for key, value := range from {
		// convert value from interface to *anypb.Any
		convertedValue, err := anypb.New(value.(proto.Message))
		if err != nil {
			log.Errorf("Unable to convert value to anypb.Any: %v", err)
			return nil
		}
		converted[key] = convertedValue
	}

	return converted

}
