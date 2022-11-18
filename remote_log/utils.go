package remote_log

import (
	"bytes"
	"encoding/binary"
	"github.com/loudbund/go-filelog/filelog_v1"
)

// UData数据转byte
func utilsEncodeUData(D []*filelog_v1.UDataSend) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	for _, d := range D {
		bytesBuffer.Write(utilsInt32ToBytes(int32(len(d.Data))))
		bytesBuffer.Write([]byte(d.Date))                     // +10=10
		bytesBuffer.Write(utilsInt32ToBytes(d.Time))          // +4=14
		bytesBuffer.Write(utilsInt64ToBytes(d.Id))            // +8=22
		bytesBuffer.Write(utilsInt16ToBytes(d.DataFileIndex)) // +2=24
		bytesBuffer.Write(utilsInt64ToBytes(d.DataOffset))    // +8=32
		bytesBuffer.Write(utilsInt16ToBytes(d.DataType))      // +2=34
		bytesBuffer.Write(utilsInt32ToBytes(d.DataLength))    // +4=38
		bytesBuffer.Write(d.Data)
	}
	return bytesBuffer.Bytes()
}

// byte数据转UData
func utilsDecodeUData(Bytes []byte) []*filelog_v1.UDataSend {
	UDatas := make([]*filelog_v1.UDataSend, 0)
	for {
		DataLength := utilsBytes2Int32(Bytes[:4])
		Bytes = Bytes[4:]
		D := &filelog_v1.UDataSend{
			Date:          string(Bytes[:10]),
			Time:          utilsBytes2Int32(Bytes[10:14]),
			Id:            utilsBytes2Int64(Bytes[14:22]),
			DataFileIndex: utilsBytes2Int16(Bytes[22:24]),
			DataOffset:    utilsBytes2Int64(Bytes[24:32]),
			DataType:      utilsBytes2Int16(Bytes[32:34]),
			DataLength:    utilsBytes2Int32(Bytes[34:38]),
			Data:          Bytes[38 : 38+DataLength],
		}
		UDatas = append(UDatas, D)
		Bytes = Bytes[DataLength+38:]

		if len(Bytes) <= 38 {
			break
		}
	}

	return UDatas
}

// 数字类型转换1： int转[]byte
func utilsInt2Bytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// 数字类型转换2： int16转[]byte
func utilsInt16ToBytes(i int16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(i))
	return buf
}

// 数字类型转换3： int32转[]byte
func utilsInt32ToBytes(i int32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

// 数字类型转换4： int64转[]byte
func utilsInt64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

// 数字类型转换5： []byte转int
func utilsBytes2Int(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	binary.BigEndian.Uint32(b)

	return int(x)
}

// 数字类型转换6： []byte转int16
func utilsBytes2Int16(buf []byte) int16 {
	return int16(binary.BigEndian.Uint16(buf))
}

// 数字类型转换7： []byte转int32
func utilsBytes2Int32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}

// 数字类型转换8： []byte转int64
func utilsBytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
