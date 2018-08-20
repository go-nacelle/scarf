# TODO - go generate instead?
# TODO - update in nacelle

secret.pb.go: secret.proto
	protoc --go_out=plugins=grpc:. $<
