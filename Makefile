PKI_PATH=$$PWD/_scratch/config
CERT_PATH=$$PWD/_scratch/certs

.PHONY: run
run:
	go run ./cmd/server/main.go

.PHONY: examples
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

.PHONY: clean
clean:
	rm ./api/v1/*.pb.go
	rm -rf ${CERT_PATH}

api/v1/log.pb.go:
	protoc api/v1/*.proto \
	                --gogo_out=\
	Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto,plugins=grpc:. \
	                --proto_path=\
	$$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
	                --proto_path=.

.PHONY: install
install: protodeps ssldeps

.PHONY: protodeps
protodeps:
	go get google.golang.org/grpc@v1.26.0
	go install github.com/gogo/protobuf/protoc-gen-gogo

.PHONY: ssldeps
ssldeps:
	go get github.com/cloudflare/cfssl/cmd/cfssl@v1.4.1
	go get github.com/cloudflare/cfssl/cmd/cfssljson@v1.4.1

_scratch/certs:
	mkdir -p ${CERT_PATH}
	cfssl gencert \
		-initca ${PKI_PATH}/ca-csr.json | cfssljson -bare ca
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${PKI_PATH}/ca-config.json \
		-profile=server \
		${PKI_PATH}/server-csr.json | cfssljson -bare server
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${PKI_PATH}/ca-config.json \
		-profile=client \
		-cn=root \
		${PKI_PATH}/client-csr.json | cfssljson -bare root
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${PKI_PATH}/ca-config.json \
		-profile=client \
		-cn=unauthorized \
		${PKI_PATH}/client-csr.json | cfssljson -bare unauthorized
	mv ca-key.pem ${CERT_PATH}
	mv ca.csr ${CERT_PATH}
	mv ca.pem ${CERT_PATH}

	mv server-key.pem ${CERT_PATH}
	mv server.csr ${CERT_PATH}
	mv server.pem ${CERT_PATH}

	mv root-key.pem ${CERT_PATH}
	mv root.pem ${CERT_PATH}
	mv root.csr ${CERT_PATH}

	mv unauthorized-key.pem ${CERT_PATH}
	mv unauthorized.pem ${CERT_PATH}
	mv unauthorized.csr ${CERT_PATH}

