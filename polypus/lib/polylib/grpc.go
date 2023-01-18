package polylib

import (
	"context"
	"encoding/hex"
	"go.dedis.ch/dela/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"math/rand"
)

func GetServerInterceptor(serverAddr string, inWatch *core.Watcher) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		//fmt.Printf("receiving - from: %s to %s\n", md.Get("FROM")[0], serverAddr)

		inWatch.Notify(Event{
			Address: md.Get("FROM")[0],
			ID:      md.Get("ID")[0],
			Msg:     "Prepare/PrePropose/Propose",
		})

		return handler(ctx, req)
	})
}

func GetDialInterceptor(clientAddr, destinationAddr string, outWatch *core.Watcher) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//fmt.Printf("sending - from %s to %s\n", clientAddr, destinationAddr)
		idBuf := make([]byte, 12)
		rand.Read(idBuf)
		ctx = metadata.AppendToOutgoingContext(ctx, "ID", hex.EncodeToString(idBuf))
		ctx = metadata.AppendToOutgoingContext(ctx, "FROM", clientAddr)

		outWatch.Notify(Event{
			Address: destinationAddr,
			ID:      hex.EncodeToString(idBuf),
			Msg:     "Promise/PreAccept/Accept",
		})

		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
