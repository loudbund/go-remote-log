package main

import (
	"fmt"
	"github.com/loudbund/go-filelog/filelog_v1"
	"github.com/loudbund/go-remote-log/read_file_log"
	"time"
)

// 6、主函数 -------------------------------------------------------------------------
func main() {
	readLog := read_file_log.NewReadLogs("/tmp/test-modsynlog-server")
	if err := readLog.Read(time.Now().Format("2006-01-02"), 0, func(Date string, DataId int64, Data *filelog_v1.UDataSend, Finish bool) int {
		// 读取结束 或 读取到空数据
		if Finish || Data == nil {
			return 0
		}
		// 打印点数据
		fmt.Println(Date, DataId, Data.DataType, string(Data.Data), Finish)
		return 1
	}); err != nil {
		fmt.Println(err)
	}
}
