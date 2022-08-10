package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"

	"BlogService-gRPC/pb"
	"BlogService-gRPC/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HTTP和gRPC不同端口的实现
// http:9090 gRPC:9002

var dPort string

func init() {
	flag.StringVar(&dPort, "port", "9002", "启动端口号")
	flag.Parse()
}

func dRungrpcserver() error {
	lis, err := net.Listen("tcp", ":"+dPort)
	if err != nil {
		return err
	}

	serv := grpc.NewServer()
	pb.RegisterTagServiceServer(serv, &server.TagServer{})
	err = serv.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

func dRunhttpserver() error {
	endpoint := "localhost:" + dPort
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, opts)
	if err != nil {
		return err
	}

	err = http.ListenAndServe(":9090", gwmux)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	//// 公钥中读取和解析公钥/私钥对
	//pair, err := tls.LoadX509KeyPair("conf/server.crt", "conf/server.key")
	//if err != nil {
	//	log.Fatalf("tls.LoadX509KeyPair err: %v", err)
	//}
	//// 创建一组根证书
	//certPool := x509.NewCertPool()
	//ca, err := ioutil.ReadFile("conf/ca.crt")
	//if err != nil {
	//	log.Fatalf("ioutil.ReadFile err: %v", err)
	//}
	//// 解析证书
	//if ok := certPool.AppendCertsFromPEM(ca); !ok {
	//	log.Fatalf("certPool.AppendCertsFromPEM err")
	//}
	//c := credentials.NewTLS(&tls.Config{
	//	Certificates: []tls.Certificate{pair},
	//	ClientAuth:   tls.RequireAndVerifyClientCert,
	//	ClientCAs:    certPool,
	//})
	//ser := grpc.NewServer(grpc.Creds(c))

	//不同端口同时起HTTP和gRPC
	errs := make(chan error)
	go func() {
		err := dRungrpcserver()
		if err != nil {
			errs <- err
		}
	}()
	go func() {
		err := dRunhttpserver()
		if err != nil {
			errs <- err
		}
	}()
	select {
	case err := <-errs:
		log.Fatalf("run server error: %v", err)
	}
}
