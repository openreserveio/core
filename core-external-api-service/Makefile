generate-oapi:
	go tool oapi-codegen -o generated/openapi/openapi.go -package openapi -generate types,client,spec,gin extapimodel/external-api.oas.yaml

generate-grpc:
	protoc --go_out=./generated --go_opt=paths=source_relative --go-grpc_out=./generated --go-grpc_opt=paths=source_relative -I ../core-ledger model/service.proto
	protoc --go_out=./generated --go_opt=paths=source_relative --go-grpc_out=./generated --go-grpc_opt=paths=source_relative -I ../core-gl glmodel/gl.proto