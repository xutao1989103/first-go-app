/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"container/list"
	"context"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/xutao1989103/first-go-app/pkg/grpc/helloworld"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) SayHelloStream(r *pb.HelloRequest, stream pb.Greeter_SayHelloStreamServer) error {
	log.Printf("Received: %v", r.GetName())
	for n := 0; n <= 6; n++ {
		log.Printf("reply times: %v", n)
		err := stream.Send(&pb.HelloReply{Message: "hello stream " + strconv.Itoa(n)})
		if err != nil {
			log.Printf("Received err: %v", err.Error())
			return err
		}
	}
	return nil
}

func (s *server) SayHelloClientStream(stream pb.Greeter_SayHelloClientStreamServer) error  {
	return nil
}

func (s *server) SayHelloBiStream(bi pb.Greeter_SayHelloBiStreamServer) error {
	nameList := list.New()
	for n := 0; n <= 3; n++ {
		request, err := bi.Recv();
		if err != nil {
			log.Printf("Received err: %v", err.Error())
			return err
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("get stream msg err:%s", err.Error())
			return nil
		}
		nameList.PushBack(request.Name)
	}

	for i:= nameList.Front(); i != nil; i=i.Next() {
		err := bi.Send(&pb.HelloReply{Message: i.Value.(string) + "-ok" })
		if err != nil {
			log.Printf("Received err: %v", err.Error())
			continue
		}
	}
	return nil;
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 5 * time.Minute,
	}))
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
