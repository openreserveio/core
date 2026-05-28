package bus

import "encoding/json"

func Encode(data interface{}) []byte {

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	return bytes

}

func Decode(data []byte, target interface{}) error {

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
