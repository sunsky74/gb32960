# gb32960

GB/T 32960《电动汽车远程服务与管理系统技术协议》第 3 部分——通信协议及数据格式的 Go 编解码库,同时支持 **2016 版与 2025 版**。模块路径 `github.com/sunsky74/gb32960`,要求 Go ≥ 1.21。

## 支持范围

- **帧层**:起始符(`##`/`$$`)、命令单元、应答标志、VIN、加密方式、数据单元长度、BCC 校验
- **V2016 报文**:车辆登入/登出、实时信息上报(整车、驱动电机、燃料电池、发动机、位置、极值、报警、可充电储能电压/温度、自定义)、平台登入/登出
- **V2025 报文**:车辆登入/登出、实时信息上报(整车、驱动电机、燃料电池发动机及车载氢系统、燃料电池电堆、发动机、位置、报警、最小并联单元电压、电池温度、超级电容、自定义、车端签名)、车辆激活/激活应答、平台登入/登出、密钥交换
- 大端字节序;异常/无效哨兵值(0xFE/0xFF 等)原样保留;解码→重编码字节级保真(真实报文金样回归验证)
- 数据单元加密仅支持 0x01 直通,RSA/AES/SM2/SM4 返回 `api.ErrEncryptionNotSupported`

## 安装

```bash
go get github.com/sunsky74/gb32960
```

## 用法

### 解码一个协议帧

```go
package main

import (
	"fmt"

	"github.com/sunsky74/gb32960/codec"
	_ "github.com/sunsky74/gb32960/codec/all" // 注册全部消息编解码器(必须)
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/utils"
)

func main() {
	raw := []byte{0x23, 0x23, 0x01, 0xFE /* ...完整一帧,含 BCC */}

	msg, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
	if err != nil {
		panic(err) // ErrInvalidHeader / ErrBCCMismatch / ErrBufferUnderflow ...
	}
	pm := msg.(*frame.ProtocolMessage)

	// 按命令把数据单元 RawBytes 解码为具体消息体(如 *gbt2016.RealTimeData)
	if err := pm.DecodePayload(); err != nil {
		panic(err)
	}
	fmt.Println(pm.Version, pm.VIN, pm.PayloadLength, pm.Payload)
}
```

### 编码一个协议帧

```go
pm := &frame.ProtocolMessage{
	Version:      api.V2016,
	RequestType:  types.CommandV2016ByCode(0x01), // 0x01 车辆登入
	ResponseType: types.ResponseCommand,          // 0xFE 命令包
	VIN:          "LSVAU2A37K2100001",
	Encryption:   types.EncryptionNone,           // 0x01 直通
	Payload:      &gbt2016.VehicleLogin{ /* ... */ },
}

frameBytes, err := pm.Bytes() // 含帧头、长度字段与 BCC
```

也可用 `codec.ProtocolCodec.Encode(w, pm)` 写入 `utils.NewByteWriter()`,输出与 `Bytes()` 逐字节一致。

### 只编解码消息体(不组帧)

```go
c := api.GetCodec(api.V2016, reflect.TypeOf(gbt2016.RealTimeData{}))

msg, err := c.Decode(utils.NewByteReader(payloadBytes)) // 解码
w := utils.NewByteWriter()
err = c.Encode(w, msg)                                  // 编码
```

### 错误处理

哨兵错误集中定义在 `api/errors.go`:`ErrInvalidHeader`、`ErrBCCMismatch`、`ErrLengthMismatch`、`ErrUnknownTLVType`、`ErrBufferUnderflow`、`ErrCodecNotFound`、`ErrEncryptionNotSupported`。编码侧对模型不一致(计数与列表长度不符、字段超长等)直接返回 error,不会产出畸形帧。

## 目录结构

```
api/     共享接口(Reader/Writer/Codecer)与错误定义
frame/   协议帧 ProtocolMessage、命令→消息体类型分发
codec/   帧层与各消息编解码器(gbt2016/、gbt2025/、all/ 汇总注册)
model/   数据模型(gbt2016/、gbt2025/)
types/   命令码、应答标志、加密方式、哨兵值等常量
utils/   大端字节流 Reader/Writer、BCC、坐标转换
golden/  真实报文金样(layer_a)
```

## 测试

```bash
go test ./...
```

覆盖:编解码往返字节级一致、金样报文重编码、哨兵值保留、编码侧一致性校验。
