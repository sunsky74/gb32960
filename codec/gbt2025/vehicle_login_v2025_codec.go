package gbt2025

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// VehicleLoginV2025Codec 编码/解码 V2025 车辆登入请求(命令 0x01)。
// 线格式与 Java VehicleLoginV2025Codec 一致:
//
//	BeanTime(6B) + SerialNum(u16) + ICCID(20B) + Count(u8)
//	+ Count × Length[u8]  (每个电池管理系统一个)
//	+ sum(Lengths) × Code[24B]
//
// 关键:每个 code 都是 24 字节,而不是每个 code Length[i] 字节(audit
// 2026-07-31)。code 总数是 sum(Lengths),而不是 Count。
type VehicleLoginV2025Codec struct{}

func init() {
	api.Register[mdl.VehicleLoginV2025](api.V2025, &VehicleLoginV2025Codec{})
}

// Decode 与 Java VehicleLoginV2025Codec.decodeBuffer 一致(第 25-53 行)。
func (c *VehicleLoginV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleLoginV2025{}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	bt, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.BeanTime = *bt.(*model.BeanTime)

	m.SerialNum = int(r.ReadUint16())
	m.ICCID = strings.TrimRight(r.ReadString(20), " ")
	m.Count = int(r.ReadUint8())

	if m.Count > 0 {
		m.Lengths = make([]int, m.Count)
		for i := 0; i < m.Count; i++ {
			m.Lengths[i] = int(r.ReadUint8())
		}

		sum := 0
		for _, l := range m.Lengths {
			sum += l
		}

		if sum > 0 {
			m.Codes = make([]string, sum)
			for i := 0; i < sum; i++ {
				m.Codes[i] = strings.TrimRight(r.ReadString(24), " ")
			}
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java VehicleLoginV2025Codec.encodeBuffer 一致(第 57-71 行)。
func (c *VehicleLoginV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleLoginV2025)

	// fix 2026-09-17: spec 2025.md L112-114(表6)—— 电池管理系统数 n(Count)
	// 0~20;每系统动力蓄电池包个数 m 0~50;动力蓄电池包编码 ∑m×24,每个为
	// 24 字节编码。Count>0 时要求 len(Lengths)==Count、len(Codes)==sum(Lengths)、
	// 每个编码 ≤24 字节;否则写入侧会静默产生与声明长度错位的帧
	// (WriteString 对超长编码会静默截断)。写入任何内容之前先校验。
	if m.Count > 0 {
		if len(m.Lengths) != m.Count {
			return fmt.Errorf("gb32960: login v2025 count %d does not match lengths length %d", m.Count, len(m.Lengths))
		}
		sum := 0
		for _, l := range m.Lengths {
			sum += l
		}
		if len(m.Codes) != sum {
			return fmt.Errorf("gb32960: login v2025 codes length %d does not match sum(lengths) %d", len(m.Codes), sum)
		}
		for i, code := range m.Codes {
			if len(code) > 24 {
				return fmt.Errorf("gb32960: login v2025 code[%d] %q longer than 24 bytes", i, code)
			}
		}
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	w.WriteString(m.ICCID, 20)
	w.WriteUint8(byte(m.Count))

	if m.Count > 0 {
		for i := 0; i < m.Count; i++ {
			w.WriteUint8(byte(m.Lengths[i]))
		}
		for _, code := range m.Codes {
			w.WriteString(code, 24)
		}
	}
	return nil
}
