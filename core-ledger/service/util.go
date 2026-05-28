package service

func ConvertMapStringToMapInterface(in map[string]string) map[string]interface{} {

	converted := make(map[string]interface{})
	for k, v := range in {
		converted[k] = v
	}

	return converted

}

func ConvertMapInterfaceToMapString(in map[string]interface{}) map[string]string {

	converted := make(map[string]string)
	for k, v := range in {
		converted[k] = v.(string)
	}

	return converted

}
