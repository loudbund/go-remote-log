package remote_log

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/loudbund/go-json/json_v1"
	"github.com/loudbund/go-remote-log/remote_log/grpc_proto_log"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	dateTypeDbLog   = int32(1001)
	dateTypeAppLog  = int32(1101)
	dateTypeJsonLog = int32(1201)
)

// 数据表日志 ------------------------------------------------------------------
type DbLog struct {
	dataType      int32
	gRpcLogClient grpc_proto_log.LogClient // grpc长连接句柄
}

// 实例化
func NewSdkDbLog(gRpcAddress string) *DbLog {
	handle := &DbLog{
		dataType: dateTypeDbLog,
	}
	// if httpDomain != "" {
	// 	handle.httpDomain = httpDomain
	// }
	if gRpcAddress != "" {
		// 创建grpc连接
		conn, err := grpc.Dial(gRpcAddress, grpc.WithInsecure())
		if err != nil {
			log.Error("日志grpc连接失败: %v", err)
		}
		// defer func() { _ = conn.Close() }()
		handle.gRpcLogClient = grpc_proto_log.NewLogClient(conn)
	}
	return handle
}

// grpc数据表日志写入
func (Me *DbLog) SdkDbLogAddGRpc(inData *grpc_proto_log.DbLogData) error {
	// 数据内容体现用proto加密
	data, err := proto.Marshal(inData)
	if err != nil {
		log.Error("marshaling error: ", err)
	}

	// 调用grpc写日志
	r, err := Me.gRpcLogClient.DbLogWrite(context.Background(), &grpc_proto_log.DbLogRequest{
		DataType: Me.dataType,
		Data:     data,
	})
	if err != nil {
		log.Error("could not greet: %v", err)
		return err
	}
	if r.ErrCode != 0 {
		return errors.New(r.ErrMessage)
	}
	return nil
}

// 应用日志 ------------------------------------------------------------------
type AppLog struct {
	dataType      int32
	gRpcLogClient grpc_proto_log.LogClient // grpc长连接句柄
}

// 实例化
func NewSdkAppLog(gRpcAddress string) *AppLog {
	handle := &AppLog{
		dataType: dateTypeAppLog,
	}
	// if httpDomain != "" {
	// 	handle.httpDomain = httpDomain
	// }
	if gRpcAddress != "" {
		// 创建grpc连接
		conn, err := grpc.Dial(gRpcAddress, grpc.WithInsecure())
		if err != nil {
			log.Error("日志grpc连接失败: %v", err)
		}
		// defer func() { _ = conn.Close() }()
		handle.gRpcLogClient = grpc_proto_log.NewLogClient(conn)
	}
	return handle
}

// grpc程序日志写入
func (Me *AppLog) SdkAppLogAddGRpc(inData *grpc_proto_log.AppLogData) error {
	// 数据内容体现用proto加密
	data, err := proto.Marshal(inData)
	if err != nil {
		log.Error("marshaling error: ", err)
	}

	// 调用grpc写日志
	r, err := Me.gRpcLogClient.AppLogWrite(context.Background(), &grpc_proto_log.AppLogRequest{
		DataType: Me.dataType,
		Data:     data,
	})
	if err != nil {
		log.Error("could not greet: %v", err)
		return err
	}
	if r.ErrCode != 0 {
		return errors.New(r.ErrMessage)
	}
	return nil
}

// Json字串格式日志 ------------------------------------------------------------------
type JsonLog struct {
	dataType      int32
	gRpcLogClient grpc_proto_log.LogClient // grpc长连接句柄
}

// 实例化
func NewSdkJsonLog(gRpcAddress string) *JsonLog {
	handle := &JsonLog{
		dataType: dateTypeJsonLog,
	}
	// if httpDomain != "" {
	// 	handle.httpDomain = httpDomain
	// }
	if gRpcAddress != "" {
		// 创建grpc连接
		conn, err := grpc.Dial(gRpcAddress, grpc.WithInsecure())
		if err != nil {
			log.Error("日志grpc连接失败: %v", err)
		}
		// defer func() { _ = conn.Close() }()
		handle.gRpcLogClient = grpc_proto_log.NewLogClient(conn)
	}
	return handle
}

// grpc数据表日志写入
func (Me *JsonLog) SdkJsonLogAddGRpc(inData interface{}) error {
	data, _ := json_v1.JsonEncode(inData)
	// 调用grpc写日志
	r, err := Me.gRpcLogClient.JsonLogWrite(context.Background(), &grpc_proto_log.JsonLogRequest{
		DataType: Me.dataType,
		Data:     []byte(data),
	})
	if err != nil {
		log.Error("could not greet: %v", err)
		return err
	}
	if r.ErrCode != 0 {
		return errors.New(r.ErrMessage)
	}
	return nil
}
