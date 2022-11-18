package remote_log

import (
	"context"
	"fmt"
	"github.com/loudbund/go-filelog/filelog_v1"
	"github.com/loudbund/go-remote-log/remote_log/grpc_proto_log"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

//
type LogServer struct {
	server *Server
}

// grpc服务实例化 Addr:0.0.0.0:1235
func NewLog(Addr string, server *Server) {
	// 创建监听服务
	listen, err := net.Listen("tcp", Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// 服务注册
	grpc_proto_log.RegisterLogServer(s, &LogServer{
		server: server,
	})

	// 启动监听
	fmt.Println("grpc开始监听:" + Addr)
	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// grpc远程函数
func (x *LogServer) DbLogWrite(ctx context.Context, in *grpc_proto_log.DbLogRequest) (*grpc_proto_log.DbLogResponse, error) { // 进行编码
	// 日志写入管道
	x.server.logChan <- &filelog_v1.UDataSend{
		DataType: int16(in.DataType),
		Data:     in.Data,
	}

	// 3、写入日志
	return &grpc_proto_log.DbLogResponse{ErrCode: 0, ErrMessage: "ok", Data: ""}, nil
}

// grpc远程函数
func (x *LogServer) AppLogWrite(ctx context.Context, in *grpc_proto_log.AppLogRequest) (*grpc_proto_log.AppLogResponse, error) { // 进行编码
	// 日志写入管道
	x.server.logChan <- &filelog_v1.UDataSend{
		DataType: int16(in.DataType),
		Data:     in.Data,
	}

	// 3、写入日志
	return &grpc_proto_log.AppLogResponse{ErrCode: 0, ErrMessage: "ok", Data: ""}, nil
}

// grpc远程函数
func (x *LogServer) JsonLogWrite(ctx context.Context, in *grpc_proto_log.JsonLogRequest) (*grpc_proto_log.JsonLogResponse, error) { // 进行编码
	// 日志写入管道
	x.server.logChan <- &filelog_v1.UDataSend{
		DataType: int16(in.DataType),
		Data:     in.Data,
	}

	// 3、写入日志
	return &grpc_proto_log.JsonLogResponse{ErrCode: 0, ErrMessage: "ok", Data: ""}, nil
}
