gen:
	protoc -I. -I ./proto --go_out=. --go-grpc_out=. --grpc-gateway_out=. --grpc-gateway_opt logtostderr=true ./proto/*.proto

clean:
	rm 	pb/*.go

run:
	go run main.go