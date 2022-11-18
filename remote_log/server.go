package remote_log

import (
	"container/list"
	"fmt"
	"github.com/loudbund/go-filelog/filelog_v1"
	"github.com/loudbund/go-json/json_v1"
	"github.com/loudbund/go-remote-log/read_file_log"
	"github.com/loudbund/go-socket2/socket2"
	"github.com/loudbund/go-utils/utils_v1"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 结构体1： 单个用户数据
type User struct {
	ClientId  string // 客户端id
	ClientIp  string // 客户端ip
	LoginTime string // App连接时间

	ReqDate   string // 请求日志的日期
	ReqDateId int64  // 请求日志的位置

	readerLog *read_file_log.ReadLogs // 日志读取工具
}

// 结构体2： 服务端结构体
type Server struct {
	SocketServer *socket2.Server
	ListUser     *list.List       // 客户端链表
	Users        map[string]*User // 客户端clientid和User的map关系
	lockListUser sync.RWMutex     // 客户端链表同步锁

	// 当前存入日志的日期
	date string

	// 日志文件目录
	logFolder string

	// 并发转线性处理通道
	logChan chan *filelog_v1.UDataSend

	// 日志处理实例map，键值为日期
	logHandles map[string]*filelog_v1.CFileLog

	// 日志保留到的天数,
	// 日志至少保留7天及最大为-7； -7，保留7天，-8，保留8天……
	retainHistoryDayNum int
}

type ParamNewServer struct {
	Ip                  string // 绑定ip
	PortSocket          int    // 数据同步用socket端口
	PortGRpc            int    // 数据接收的grpc端口
	LogFolder           string // 日志文件夹
	RetainHistoryDayNum int    // 日志保留天数
	SendFlag            int    // socket验证标记
}

// 对外函数：创建实例
func NewServer(Param ParamNewServer) *Server {
	Me := &Server{
		Users:      map[string]*User{},
		ListUser:   list.New(),
		date:       utils_v1.Time().Date(),
		logFolder:  Param.LogFolder,
		logChan:    make(chan *filelog_v1.UDataSend),
		logHandles: map[string]*filelog_v1.CFileLog{},
	}
	// -1- 参数校正
	if Param.RetainHistoryDayNum > -8 {
		Me.retainHistoryDayNum = -8
		fmt.Println("Warning! retainHistoryDayNum is big than -8,will use -8")
	} else {
		Me.retainHistoryDayNum = Param.RetainHistoryDayNum
	}

	// -2- 关闭前几天的日志
	Me.closePreDateLog()

	// -3- 写日志协程
	go Me.messageWrite()

	// -4- socket服务器
	Me.SocketServer = socket2.NewServer(Param.Ip, Param.PortSocket, func(Event socket2.HookEvent) {
		Me.onHookEvent(Event)
	})

	// -5- grpc服务，
	if Param.PortGRpc > 0 {
		go NewLog(Param.Ip+":"+strconv.Itoa(Param.PortGRpc), Me)
	}

	// -6- 校验码设置
	Me.SocketServer.Set("SendFlag", Param.SendFlag)

	return Me
}

// 接收直接发过来的数据
func (Me *Server) CommitData(DataType int16, Data []byte) {
	Me.logChan <- &filelog_v1.UDataSend{
		DataType: DataType,
		Data:     Data,
	}
}

// 关闭前几天的日志文件
func (Me *Server) closePreDateLog() {
	Today := utils_v1.Time().Date()
	iDate := utils_v1.Time().DateAdd(utils_v1.Time().Date(), -7)
	I := 0
	for {
		// 最多执行20次，避免死循环
		I++
		if I > 20 {
			break
		}

		// 日期已经是今天了，停止处理
		if iDate == Today {
			break
		}

		// 处理这天的
		handle := filelog_v1.New(Me.logFolder, iDate)
		handle.SetFinish()
		handle.Close()

		// 日期加1天
		iDate = utils_v1.Time().DateAdd(iDate, 1)
	}

	// 执行一次日志清理
	Me.logDelete()
}

// 管道接收日志并写文件
func (Me *Server) messageWrite() {
	// 日志handle不存在则先创建日志handle
	if _, ok := Me.logHandles[Me.date]; !ok {
		Me.logHandles[Me.date] = filelog_v1.New(Me.logFolder, Me.date)
	}
	// 定时器设置
	T := time.NewTicker(time.Second)
	for {
		select {
		case <-T.C:
			// 每秒都判断刷新下Me.date 和 日志handle
			// 判断是否跨天了，跨天需要删除当前的日志句柄，并修改当前的日志日期Me.date
			Time := time.Now().Unix()
			if true {
				Date := utils_v1.Time().Date(time.Unix(Time, 0))
				if Date != Me.date {
					// 关闭
					if _, ok := Me.logHandles[Me.date]; ok {
						Me.logHandles[Me.date].SetFinish() // 设置finish标记
						Me.logHandles[Me.date].Close()
						delete(Me.logHandles, Me.date)
					}
					Me.date = Date
					// 日志handle不存在则先创建日志handle
					if _, ok := Me.logHandles[Me.date]; !ok {
						Me.logHandles[Me.date] = filelog_v1.New(Me.logFolder, Me.date)
					}
				}
			}

		case D, ok := <-Me.logChan:
			if !ok {
				return
			}
			// fmt.Println(D.DataType, string(D.Data))
			// 判断是否跨天了，跨天需要删除当前的日志句柄，并修改当前的日志日期Me.date
			Time := time.Now().Unix()
			if true {
				Date := utils_v1.Time().Date(time.Unix(Time, 0))
				if Date != Me.date {
					// 关闭
					if _, ok := Me.logHandles[Me.date]; ok {
						Me.logHandles[Me.date].SetFinish() // 设置finish标记
						Me.logHandles[Me.date].Close()
						delete(Me.logHandles, Me.date)
					}
					Me.date = Date

					// 执行一次日志清理
					Me.logDelete()
				}
			}

			// 日志handle不存在则先创建日志handle
			if _, ok := Me.logHandles[Me.date]; !ok {
				Me.logHandles[Me.date] = filelog_v1.New(Me.logFolder, Me.date)
			}
			// 写入到文件日志
			if _, err := Me.logHandles[Me.date].Add(int32(Time), D.DataType, D.Data); err != nil {
				log.Error(err)
			}
		}
	}
}

// //////////////////////////////////////////////////////////////////////
// socket日志同步服务 -------------------------
// //////////////////////////////////////////////////////////////////////

// 1、处理数据,多线程转单线程处理
func (Me *Server) onHookEvent(Event socket2.HookEvent) {
	switch Event.EventType {
	case "message": // 1、消息事件
		// 客户端请求日期和开始位置日志
		if Event.Message.CType == 301 {
			if jData, err := json_v1.JsonDecode(string(Event.Message.Content)); err != nil {
				log.Error(err)
			} else {
				Date, err1 := json_v1.GetJsonString(jData, "date")
				id, err2 := json_v1.GetJsonInt64Force(jData, "id")
				if err1 != nil {
					log.Error(err1)
				} else if err2 != nil {
					log.Error(err2)
				} else {
					Me.lockListUser.Lock()
					if _, ok := Me.Users[Event.User.ClientId]; ok {
						Me.Users[Event.User.ClientId].ReqDate = Date
						Me.Users[Event.User.ClientId].ReqDateId = id
					}
					Me.lockListUser.Unlock()
				}
			}
		}

	case "offline": // 2、下线事件
		Me.removeUser(Event.User.ClientId)

	case "online": // 3、上线消息
		U := Me.addUser(Event.User.ClientId, Event.User.Addr)
		go Me.sendLog(U)
	}
}

// 批量获取log日志
func (Me *Server) getLogGroup(U *User, rowNumber int) ([]*filelog_v1.UDataSend, bool) {
	Me.lockListUser.Lock()
	KeyDate := Me.Users[U.ClientId].ReqDate
	KeyDateId := Me.Users[U.ClientId].ReqDateId
	Me.lockListUser.Unlock()

	KeyData := make([]*filelog_v1.UDataSend, 0)
	KeyFinish := false
	if err := U.readerLog.Read(KeyDate, KeyDateId, func(Date string, DataId int64, Data *filelog_v1.UDataSend, Finish bool) int {
		if Finish || Date != KeyDate { // 日期参数的日志已读取完毕
			KeyFinish = true
			return 0 // 终止读取
		}
		if Data == nil {
			return 0 // 终止读取
		}
		KeyData = append(KeyData, Data)
		if len(KeyData) >= rowNumber {
			return 0 // 终止读取
		}
		return 1 // 继续读取
	}); err != nil {
		log.WithFields(log.Fields{"n": "取数据失败"}).Error(err)
	}
	return KeyData, KeyFinish
}

// 发送日志给客户端
func (Me *Server) sendLog(U *User) {
	fmt.Println("start send log:", U.ClientId)
	for {
		Me.lockListUser.Lock()
		_, ok := Me.Users[U.ClientId]
		Me.lockListUser.Unlock()
		if !ok {
			return
		}

		if U.ReqDate != "" {
			KeyPerNum := 500
			KeyData, KeyFinish := Me.getLogGroup(U, KeyPerNum)
			// 打印点输出
			if len(KeyData) > 0 {
				fmt.Println(utils_v1.Time().DateTime(), "send log to ", U.ClientId, len(KeyData))
			}

			// 有数据需要处理
			if len(KeyData) > 0 {
				if err := Me.SocketServer.SendMsg(&U.ClientId, socket2.UDataSocket{
					Zlib:    1,
					CType:   302,
					Content: utilsEncodeUData(KeyData),
				}); err != nil {
					log.WithFields(log.Fields{"n": "消息发送失败"}).Error(err)
					return
				} else {
					Me.lockListUser.Lock()
					if _, ok := Me.Users[U.ClientId]; ok {
						Me.Users[U.ClientId].ReqDateId += int64(len(KeyData))
					}
					Me.lockListUser.Unlock()
				}
				if len(KeyData) == KeyPerNum {
					continue
				}
			}
			// 如果日期内的日志已经发送完成，则发送标记
			if KeyFinish {
				fmt.Println(utils_v1.Time().DateTime(), "日期日志发送结束 ", U.ClientId, U.ReqDate)
				if err := Me.SocketServer.SendMsg(&U.ClientId, socket2.UDataSocket{
					Zlib:    1,
					CType:   304,
					Content: []byte(U.ReqDate),
				}); err != nil {
					log.WithFields(log.Fields{"n": "消息发送失败"}).Error(err)
					return
				}
				// 停止扫描发送
				U.ReqDate = ""
			}
			time.Sleep(time.Second)
			// }
		} else {
			time.Sleep(time.Second)
		}
	}
}

// 添加用户
func (Me *Server) addUser(ClientId, Addr string) *User {
	IpPort := strings.Split(Addr, ":")
	U := &User{
		ClientId:  ClientId,
		ClientIp:  IpPort[0],
		LoginTime: utils_v1.Time().DateTime(),
		readerLog: read_file_log.NewReadLogs(Me.logFolder),
	}
	Me.lockListUser.Lock()
	Me.ListUser.PushBack(U)
	Me.Users[U.ClientId] = U
	Me.lockListUser.Unlock()

	return U
}

// 移除用户
func (Me *Server) removeUser(ClientId string) {
	Me.lockListUser.Lock()
	for e := Me.ListUser.Front(); e != nil; e = e.Next() {
		if e.Value.(*User).ClientId == ClientId {
			e.Value.(*User).readerLog.Close()
			Me.ListUser.Remove(e)
			delete(Me.Users, ClientId)
		}
	}
	Me.lockListUser.Unlock()
}

// 3、日志清理
func (Me *Server) logDelete() {
	// 开始保留日志的日期
	retainDate := utils_v1.Time().DateAdd(utils_v1.Time().Date(), Me.retainHistoryDayNum)

	// 从需要保留的日期，向前删除30天的数据
	Start := utils_v1.Time().DateAdd(retainDate, -30)
	for D := Start; D < retainDate; D = utils_v1.Time().DateAdd(D, 1) {
		Folder := Me.logFolder + "/" + strings.ReplaceAll(D, "-", "")
		if utils_v1.File().CheckFileExist(Folder) {
			log.Info("清理日志 " + Folder)
			_ = os.RemoveAll(Folder)
		}
	}
}
