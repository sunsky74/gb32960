package model

import "fmt"

// BeanTime 表示线格式的数据采集时间戳。
// Year 存放线格式的偏移值:Year = actualYear - 2000。
// 例如:Year=0x01 表示 2001 年,Year=0x25 表示 2037 年。
type BeanTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

// String 返回人类可读的表示。
func (bt BeanTime) String() string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		bt.Year+2000, bt.Month, bt.Day, bt.Hour, bt.Minute, bt.Second)
}
