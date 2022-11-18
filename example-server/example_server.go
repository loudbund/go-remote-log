package main

import (
	"github.com/loudbund/go-remote-log/remote_log"
)

// 6、主函数 -------------------------------------------------------------------------
func main() {
	// 1、创建客户端连接
	s := remote_log.NewServer(remote_log.ParamNewServer{
		Ip:                  "127.0.0.1",
		PortSocket:          2222,
		PortGRpc:            1235,
		LogFolder:           "/tmp/test-modsynlog-server",
		RetainHistoryDayNum: -5,
		SendFlag:            123456,
	})

	// 2、直接发一条数据过去
	s.CommitData(1234, []byte("haha12345678"))

	// 处理其他逻辑
	select {}
}
