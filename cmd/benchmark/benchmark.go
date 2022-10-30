package benchmark

import (
	"fmt"
	"time"

	"github.com/guonaihong/gout"
)

type Benchmark struct {
	Server     string        `clop:"short;long" usage:"server address" valid:"required"`
	Put        bool          `clop:"short;long" usage:"create file"`
	Concurrent int           `clop:"short;long" usage:"Concurrent" default:"10"`
	Number     int           `clop:"short;long" usage:"number"`
	Body       string        `clop:"short;long" usage:"body"`
	Durations  time.Duration `clop:"short;long" usage:"duration"`
}

func (b *Benchmark) put() {

	err := gout.
		POST(b.Server + "/file/raw"). //压测本地8080端口
		SetBody(b.Body).              //设置请求body内容
		Filter().                     //打开过滤器
		Bench().                      //选择bench功能
		Concurrent(b.Concurrent).     //并发数
		Number(b.Number).             //压测次数
		Durations(b.Durations).
		Do()

	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func (b *Benchmark) SubMain() {
	if b.Put {
		b.put()
	}
}
