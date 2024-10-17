.PHONY: protos

protos:
	protoc -I protos/ --go_out=protos --go-grpc_out=require_unimplemented_servers=false:protos protos/currency.proto
