generate-grpc:
	protoc --go_out=./generated --go_opt=paths=source_relative --go-grpc_out=./generated --go-grpc_opt=paths=source_relative -I ../core-ledger model/service.proto

dockerize:
	docker build -f application/coreledgerposter/docker/Dockerfile -t finsorb/coreledgerposter .