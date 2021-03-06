package account

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/xiaomeng79/go-log"
	"github.com/xiaomeng79/istio-micro/cinit"
	"github.com/xiaomeng79/istio-micro/internal/gateway"
	"github.com/xiaomeng79/istio-micro/internal/wrapper"
	pb "github.com/xiaomeng79/istio-micro/srv/account/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

const (
	SN = "srv-account" //定义services名称
)

func Run() {

	//初始化,选着需要的组件
	cinit.InitOption(SN, "trace", "postgres", "metrics")
	// 更新sql
	err := execUpdateSql()
	if err != nil {
		log.Fatalf("updatesql error:%+v ", err)
	}
	// 更新版本号
	err = updateVersion()
	if err != nil {
		log.Fatalf("update version error:%+v ", err)
	}
	lis, err := net.Listen("tcp", cinit.Config.SrvAccount.Port)
	if err != nil {
		log.Fatal("failed to listen: " + err.Error())
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_opentracing.UnaryServerInterceptor(),
			wrapper.RecoveryUnaryInterceptor,
			wrapper.LoggingUnaryInterceptor,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_opentracing.StreamServerInterceptor(),
		)),
	)
	pb.RegisterAccountServiceServer(s, &AccountServer{})
	reflection.Register(s)

	// 开启一个网关服务
	ctx := context.Background()
	go gateway.Run(
		ctx,
		gateway.WithAddr(cinit.Config.SrvAccount.GateWayAddr),
		gateway.WithGRPCServer("tcp", cinit.Config.SrvAccount.Address),
		gateway.WithSwaggerDir(cinit.Config.SrvAccount.GateWaySwaggerDir),
		gateway.WithHandle(pb.RegisterAccountServiceHandler),
	)

	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to listen: " + err.Error())
	}
}
