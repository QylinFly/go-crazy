package Dubbo

import (
  "strconv"
  "encoding/binary"
  "bytes"
"github.com/xoxo/crm-x/util/logger"
 "sync/atomic"
)

const (
    // header length.
	HEADER_LENGTH = 16
	// magic header.
	 

    // message flag.
    FLAG_REQUEST byte= 0x80
    FLAG_TWOWAY byte= 0x40
	FLAG_EVENT byte= 0x20
	

	// Magic - Magic High & Magic Low (16 bits)
	// Identifies dubbo protocol with value: 0xdabb.
	MAGIC uint16 = 0xdabb
	
	MAGIC_H byte = 0xda
	MAGIC_L byte = 0xbb

	// Req/Res (1 bit)
	// Identifies this is a request or response. Request - 1; Response - 0.
	// 2 Way (1 bit)
	// Only useful when Req/Res is 1 (Request), expect for a return value from server or not. Set to 1 if need a return value from server.
	// Event (1 bit)
	// Identifies an event message or not, for example, heartbeat event. Set to 1 if this is an event.
	// Serialization ID (5 bit)
	// Identifies serialization type: the value for fastjson is 6.

	//Req-notEvent-fastjson --    11000110
	REQ_FASTJSON byte = 0xc6
	//Res-notEvent-fastjson --    01000110
	RES_FASTJSON byte = 0x46

	// Status (8 bits)
	// Only useful when Req/Res is 0 (Response), identifies the status of response:
	// 20 - OK
	// 30 - CLIENT_TIMEOUT
	// 31 - SERVER_TIMEOUT
	// 40 - BAD_REQUEST
	// 50 - BAD_RESPONSE
	// 60 - SERVICE_NOT_FOUND
	// 70 - SERVICE_ERROR
	// 80 - SERVER_ERROR
	// 90 - CLIENT_ERROR
	// 100 - SERVER_THREADPOOL_EXHAUSTED_ERROR
	STATUS byte = 0x00 //20

	
	// Request ID (64 bits)
	// Identifies an unique request. Numeric (long).

	// Data Length (32)
	// Length of the content (the variable part) after serialization, counted by bytes. Numeric (integer).

	// Variable Part
	// Each part is a byte[] after serialization with specific serialization type, identifies by Serialization ID.

	// Every part is a byte[] after serialization with specific serialization type, identifies by Serialization ID.

	// If the content is a Request (Req/Res = 1), each part consists of the content, in turn is:
)

type DubboRpcEncoder struct{
	buf  	*bytes.Buffer
	header 	*bytes.Buffer
	reqId 	uint64
}

func New() *DubboRpcEncoder{
	dubbo := &DubboRpcEncoder{
		buf 	: new(bytes.Buffer),
		header 	: new(bytes.Buffer),
		reqId	: 0,
	}
	dubbo.InitHeader()
	return dubbo
}

func (dubbo * DubboRpcEncoder) InitHeader() *DubboRpcEncoder{
	// dubbo.header.
	binary.Write(dubbo.header, binary.BigEndian, MAGIC)
	dubbo.header.WriteByte(REQ_FASTJSON)
	dubbo.header.WriteByte(STATUS)
	return dubbo
}

/*
	"2.0.1"
	"com.alibaba.dubbo.performance.demo.provider.IHelloService"
	null
	"hash"
	"Ljava/lang/String;"
	"fff"
	{"path":"com.alibaba.dubbo.performance.demo.provider.IHelloService"}
*/

// String interfaceName, String method, String parameterTypesString, String parameter
func (dubbo * DubboRpcEncoder) GetEncoderData(parameter string) (uint64,*bytes.Buffer){

	body := new(bytes.Buffer)
	data := new(bytes.Buffer)

	data.WriteString("\"2.0.1\"\n")
	data.WriteString("\"com.alibaba.dubbo.performance.demo.provider.IHelloService\"\n")
	data.WriteString("null\n")
	data.WriteString("\"hash\"\n")
	data.WriteString("\"Ljava/lang/String;\"\n")
	data.WriteString("\"" + parameter + "\"\n")
	data.WriteString("{\"path\":\"com.alibaba.dubbo.performance.demo.provider.IHelloService\"}\n")
	
	reqId :=atomic.AddUint64(&dubbo.reqId,1)
	length := data.Len()
	body.Write(dubbo.header.Bytes())
	binary.Write(body, binary.BigEndian, reqId)
	binary.Write(body, binary.BigEndian, int32(length))

	body.Write(data.Bytes())
	return reqId,body
}

//  返回解析结构体
type RpcResponse struct{
	 requestId uint64
	 bytes []byte
}

func (res* RpcResponse)ID() uint64{
	return res.requestId
}
func (res* RpcResponse)Data() *[]byte{
	return &res.bytes
}


func  GetDecoderData(data *[]byte,res []*RpcResponse) (v []*RpcResponse, errstr string, last *[]byte){

	for {
		len := len(*data)
			
		if (len < 2) {
			return res, "数据字段长度小于2错误",data
		}

		dubboByte :=  (*data)[0:2]
		dubbo := binary.BigEndian.Uint16(dubboByte)
		if dubbo != MAGIC{
			return res, "头部校验失败！",data
		}

		if (len < HEADER_LENGTH) {
			return res, "数据字段长度小于头部最小长度16",data
		}

		statusByte := (*data)[3]
		status := int8(statusByte)
		if status != 20{
			// 20 - OK
			// 30 - CLIENT_TIMEOUT
			// 31 - SERVER_TIMEOUT
			// 40 - BAD_REQUEST
			// 50 - BAD_RESPONSE
			// 60 - SERVICE_NOT_FOUND
			// 70 - SERVICE_ERROR
			// 80 - SERVER_ERROR
			// 90 - CLIENT_ERROR
			// 100 - SERVER_THREADPOOL_EXHAUSTED_ERROR
			logger.Info("数据状态异常 code="+ strconv.Itoa(int(status)))
			// return res, "数据状态异常 code="+ strconv.Itoa(int(status)),nil
		}
		
		dataLenByte :=  (*data)[12:16] 
		dataLen :=binary.BigEndian.Uint32(dataLenByte)
		tt := int(dataLen) + HEADER_LENGTH
		if (len < tt) {
			return res, "头部加数据小于数据长度",data
		}

		if status == 20{
			requestIdBytes := (*data)[4:12]
			requestId :=binary.BigEndian.Uint64(requestIdBytes)
		
			rpcResponse := &RpcResponse{
				requestId : requestId,
				bytes : (*data)[HEADER_LENGTH+2:tt-1] ,
			}
			res = append(res,rpcResponse)
		}

		if (len > tt) {
			// return res, "严重错误，可能粘包了"
			dd := (*data)[tt:len]
			data = &dd
			// return GetDecoderData(&dd,res)
		}else{
			return res, "",nil
		}

	}
}




// ////////////


// // Version of this protocol
// const ProtocolVersion = uint8(1)

// // Protocol message types
// const (
//   MsgTypeSingleReq     = MsgType('r')
//   MsgTypeStreamReq     = MsgType('s')
//   MsgTypeStreamReqPart = MsgType('p')
//   MsgTypeSingleRes     = MsgType('R')
//   MsgTypeStreamRes     = MsgType('S')
//   MsgTypeErrorRes      = MsgType('E')
//   MsgTypeRetryRes      = MsgType('e')
//   MsgTypeNotification  = MsgType('n')
//   MsgTypeHeartbeat     = MsgType('h')
//   MsgTypeProtocolError = MsgType('f')
// )

// // ProtocolError codes
// const (
//   ProtocolErrorAbnormal    = 0
//   ProtocolErrorUnsupported = 1
//   ProtocolErrorInvalidMsg  = 2
//   ProtocolErrorTimeout     = 3
// )


// // Protocol message type
// type MsgType byte

// // Write the version this protocol implements to `s`
// func WriteVersion(s io.Writer) (int, error) {
//   return s.Write(protocolVersionBuf[:])
// }

// // Read the version the other end implements. Returns an error if this side's protocol
// // is incompatible with the other side's version.
// func ReadVersion(s io.Reader) (uint8, error) {
//   b := make([]byte, 2)
//   if _, err := readn(s, b); err != nil {
//     return 0, err
//   }
//   n, err := strconv.ParseUint(string(b), 16, 8)
//   if err != nil {
//     return 0, err
//   }
//   if n != uint64(ProtocolVersion) {
//     return 0, errors.New("unsupported protocol version \"" + string(b) + "\"")
//   }
//   return uint8(n), nil
// }


// // Maximum value of a heartbeat's "load"
// var HeartbeatMsgMaxLoad = 0xffff


// // Create a slice of bytes representing a heartbeat message
// func MakeHeartbeatMsg(load uint16) []byte {
//   b := []byte{byte(MsgTypeHeartbeat),0,0,0,0,0,0,0,0,0,0,0,0}
//   z := 1
//   copyFixnum(b[z:z+4], 4, uint64(load), 16)
//   z += 4
//   copyFixnum(b[z:z+8], 8, uint64(time.Now().UTC().Unix()), 16)
//   return b
// }


// // Create a slice of bytes representing a message (w/o any payload)
// func MakeMsg(t MsgType, id, name3 string, wait, size int) []byte {
//   // calculate buffer size
//   bz := 9  // minimum size, fitting type and payload size
//   name3z := 0

//   if t == MsgTypeRetryRes {
//     bz = 21  // e.g. "e00010000000100000001"
//   } else {
//     if id != "" {
//       bz += 4  // msg with id e.g. "R000100000005"
//     }
//     name3z = len(name3)
//     if name3z != 0 {
//       bz += 3 + name3z  // msg w/ name3 e.g. "r0001004echo00000005"
//     }
//   }


//   b := make([]byte, bz)
//   b[0] = byte(t)  // type e.g. "R"
//   z := 1

//   if id != "" {
//     b[1] = id[0]
//     b[2] = id[1]
//     b[3] = id[2]
//     b[4] = id[3]  // id e.g. "abcd"
//     z += 4
//   }

//   if name3z != 0 {
//     if len(name3) == 0 {
//       panic("empty name")
//     }
//     copyFixnum(b[z:z+3], 3, uint64(name3z), 16) // name3 size e.g. "004"
//     z += 3
//     copy(b[z:], []byte(name3))
//     z += name3z
//   }

//   if t == MsgTypeRetryRes {
//     if wait == 0 {
//       copy(b[z:], zeroes[:8])
//     } else {
//       copyFixnum(b[z:z+8], 8, uint64(wait), 16)
//     }
//     z += 8
//   }

//   if size == 0 {
//     copy(b[z:], zeroes[:8])
//   } else {
//     copyFixnum(b[z:z+8], 8, uint64(size), 16)  // payload size e.g. "0000005"
//   }

//   return b[:z+8]
// }


// // Read a message from `s`
// // If t is MsgTypeHeartbeat, wait==load, size==time
// func ReadMsg(s io.Reader) (t MsgType, id, name3 string, wait, size uint32, err error) {
//   // "r0001004echo00000005"  => ('r', "0001", "echo", 0, 5, nil)
//   // "R000100000005"         => ('R', "0001", "", 0, 5, nil)
//   // "e00010000138800000014" => ('e', "0001", "", 5000, 20, nil)
//   b := make([]byte, 128)

//   // A message has a minimum size of 13, so read first 13 bytes
//   // e.g. "n001a00000000" = <notification> <short name> <no payload>
//   readz := 13
//   readz, err = readn(s, b[:readz])
//   if err != nil {
//     if err == io.EOF && readz >= 9 && b[0] == byte(MsgTypeProtocolError) {
//       // OK to read until EOF for MsgTypeProtocolError as they are shorter than other messages
//       err = nil
//     } else {
//       return
//     }
//   }

//   // type
//   t = MsgType(b[0])
//   z := 1

//   if t == MsgTypeHeartbeat {
//     // load
//     var n uint64
//     n, err = strconv.ParseUint(string(b[z:z+4]), 16, 16)
//     z += 4
//     if err != nil {
//       return
//     }
//     wait = uint32(n)

//   } else if t != MsgTypeNotification && t != MsgTypeProtocolError {
//     // requestID
//     id = string(b[z:z+4])
//     z += 4
//   }

//   if t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification {
//     // name
//     // text3Size
//     name3z, e := strconv.ParseUint(string(b[z:z+3]), 16, 16)
//     z += 3
//     if e != nil {
//       err = e
//       return
//     }

//     // Read remainder of message
//     newz := z + int(name3z) + 8  // 8 = payload size

//     if cap(b) < newz {
//       // Grow buffer (only happens with really long name3)
//       newb := make([]byte, newz)
//       copy(newb, b)
//       b = newb
//     }

//     if newz > readz {
//       if _, err = readn(s, b[readz:newz]); err != nil {
//         return
//       }
//     }

//     // text3Value
//     name3 = string(b[z:z+int(name3z)])
//     z += int(name3z)

//   } else if t == MsgTypeRetryRes {
//     // wait
//     n, e := strconv.ParseUint(string(b[z:z+8]), 16, 32)
//     if e != nil {
//       err = e
//       return
//     }
//     wait = uint32(n)
//     z += 8
//     // read remainding 8 bytes of the message
//     if _, err = readn(s, b[z:z+8]); err != nil {
//       return
//     }
//   }

//   // payloadSize (or time if t==MsgTypeHeartbeat)
//   n, e := strconv.ParseUint(string(b[z:z+8]), 16, 32)
//   if e != nil {
//     err = e
//     return
//   }
//   size = uint32(n)

//   return
// }


// // Returns a 4-byte representation of a 32-bit integer, suitable an integer-based request ID.
// func FormatRequestID(n uint32) []byte {
//   buf := bytes.NewBuffer(make([]byte,4)[:0])
//   err := binary.Write(buf, binary.LittleEndian, n)
//   if err != nil {
//     panic(err)
//   }
//   return buf.Bytes()
// }


// // =============================================================================

// var protocolVersionBuf [2]byte  // version as fixnum2 ([H,H] where H is any of 0-9a-fA-F)
// var zeroes = [...]byte{48,48,48,48,48,48,48,48}

// func init() {
//   copyFixnum(protocolVersionBuf[:0], 2, uint64(ProtocolVersion), 16)
// }


// func copyFixnum(buf []byte, ndigits int, n uint64, base int) {

	

//   z := len(strconv.AppendUint(buf[:0], n, base))
//   rShiftSlice(buf[:ndigits], ndigits-z, byte(48))
// }


// func makeFixnumBuf(ndigits int, n uint64, base int) []byte {
//   zb := make([]byte, ndigits)
//   z := len(strconv.AppendUint(zb[:0], n, base))
//   rShiftSlice(zb, ndigits-z, byte(48))
//   return zb
// }


// func rShiftSlice(b []byte, n int, padb byte) {
//   if n != 0 {
//     bz := len(b)
//     wi := bz-1
//     for ; wi >= n; wi-- {
//       b[wi] = b[wi-n]
//       b[wi-n] = padb
//     }
//     zn := bz-n-1
//     for ; wi > zn; wi-- {
//       b[wi] = padb
//     }
//   }
// }


// // Read exactly len(b) bytes from s, blocking if needed
// func readn(s io.Reader, b []byte) (int, error) {
//   // behaves similar to io.ReadFull, but simpler and allowing EOF<len(b)
//   p := 0
//   n := len(b)
//   for p < n {
//     z, err := s.Read(b[p:])
//     p += z
//     if err != nil {
//       return p, err
//     }
//   }
//   return p, nil
// }