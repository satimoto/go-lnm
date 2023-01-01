protoc lsprpc/channel.proto --go_out=plugins=grpc:$GOPATH/src
protoc lsprpc/invoice.proto --go_out=plugins=grpc:$GOPATH/src
