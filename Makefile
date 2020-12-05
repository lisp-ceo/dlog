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
