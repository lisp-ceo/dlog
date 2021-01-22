run:
	go run ./cmd/server/main.go

examples:
	curl -X POST localhost:8888 -d \
	    '{"record": {"value": "JMX0J3MgR28gIzEK"}}'
	curl -X POST localhost:8888 -d \
	    '{"record": {"value": "JMX0J3MgR28gIzIK"}}'
	curl -X POST localhost:8888 -d \
	    '{"record": {"value": "JMX0J3MgR28gIzMK"}}'
	curl -X GET localhost:8888 -d '{"offset": 0}'
	curl -X GET localhost:8888 -d '{"offset": 1}'
	curl -X GET localhost:8888 -d '{"offset": 2}'

.PHONY: build
build: api/v1/log.pb.go

clean:
	rm ./api/v1/*.pb.go

api/v1/log.pb.go:
	protoc api/v1/*.proto \
	                --gogo_out=\
	Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto,plugins=grpc:. \
	                --proto_path=\
	$$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
	                --proto_path=.

install:
	go get google.golang.org/grpc@v1.26.0
	go install github.com/gogo/protobuf/protoc-gen-gogo