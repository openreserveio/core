generate-grpc:
	protoc --go_out=./generated --go_opt=paths=source_relative --go-grpc_out=./generated --go-grpc_opt=paths=source_relative model/service.proto


dockerize:
	docker build -f application/coreledger/docker/Dockerfile -t openreserveio/coreledger .