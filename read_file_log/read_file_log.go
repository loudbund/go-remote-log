// 1、使用需要先执行 NewReadLogs 函数，参数为日志文件夹
// 2、读取日志调用 Read(DateStart string, DataIdStart int64, CallBackData func(Date string, DataId int64, Data *filelog_v1.UDataSend, Finish bool) int) error {
// **** 日期日志文件夹里没有 finish标记时，日期内的日志读完后也会一直停留在这个日期里 ****
package read_file_log

import (
	"fmt"
	"github.com/loudbund/go-filelog/filelog_v1"
	"github.com/loudbund/go-utils/utils_v1"
	"time"
)

// 位置记录结构体
type ReadLogs struct {
	simpleLogFolder string // 日志目录

	Date   string // 日期
	DataId int64  // 下一数据id

	fileLogHandle *filelog_v1.CFileLog // 文件日志类句柄
	closed        bool                 // 关闭状态
}

// 获取句柄
func NewReadLogs(SimpleLogFolder string) *ReadLogs {
	// 实例化
	handle := &ReadLogs{
		simpleLogFolder: SimpleLogFolder,
		fileLogHandle:   nil,
		closed:          false,
	}
	return handle
}

// 关闭
func (Me *ReadLogs) Close() {
	Me.closed = true
}

// 切换到新一天日志
func (Me *ReadLogs) changeDate(Date string, DataId int64) {
	// 默认数据
	Me.DataId = DataId
	// 日期有变化时，日志句柄重新生成
	if Me.Date != Date {
		Me.Date = Date
		if Me.fileLogHandle != nil {
			Me.fileLogHandle.Close()
		}
		fmt.Println("【read_logs】切换日期到", Me.Date, Me.DataId)
		Me.fileLogHandle = filelog_v1.New(Me.simpleLogFolder, Me.Date)
	}
}

// 位置回调
func (Me *ReadLogs) callbackPos(CallbackPos func(Date string, DataId int64)) {
	// 位置回调
	if CallbackPos != nil {
		CallbackPos(Me.Date, Me.DataId)
	}
}

// 阻塞方式读取一条日志
// CallBackData 返回 1：成功，2：重试，0：结束
func (Me *ReadLogs) Read(DateStart string, DataIdStart int64, CallBackData func(Date string, DataId int64, Data *filelog_v1.UDataSend, Finish bool) int) error {
	Me.changeDate(DateStart, DataIdStart)

	// 循环处理
	for {
		// 已标记为关闭了
		if Me.closed {
			if Me.fileLogHandle != nil {
				Me.fileLogHandle.Close()
			}
			return nil
		}
		// 读取一条数据
		if D, err := Me.fileLogHandle.GetOne(Me.DataId); err != nil { // 读取出错
			fmt.Println(err)
			return err
		} else if D != nil { // 读取到的数据不为nil
			// 回调,
			status := CallBackData(Me.Date, Me.DataId, D, false)
			if status == 1 { // 成功
				Me.DataId++
			} else if status == 2 { // 重试不刷新位置
				fmt.Println("【read_logs】重试不刷新位置", Me.Date, Me.DataId)
				time.Sleep(time.Second)
			} else if status == 0 {
				// 退出
				return nil
			}
		} else { // 读取到nil数据
			// 判断这个日期数据是否已全部读取
			Finish := Me.fileLogHandle.GetFinish(true)
			if CallBackData(Me.Date, Me.DataId, nil, Finish) == 0 {
				return nil
			}
			// 已经读完并且当前日志不是今天，则切换到下一天读取
			if Finish && Me.Date < utils_v1.Time().Date() {
				Me.changeDate(utils_v1.Time().DateAdd(Me.Date, 1), 0)
			}
			time.Sleep(time.Second)
		}
	}
}
