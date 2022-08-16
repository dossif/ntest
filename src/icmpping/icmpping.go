package icmpping

import (
	"context"
	"fmt"
	"pping/src/args"
	"time"
)

type Test struct {
}

//type Ping struct {
//	Host     string
//	Ip       string
//	Sequence int
//	Status   bool
//	Rtt      time.Duration
//	Error    error
//}
//
//type Pinger struct {
//	Api *goping.Pinger
//	Log *log.Logger
//	Ctx context.Context
//	Wg  *sync.WaitGroup
//}

func NewTest() *Test {
	return &Test{}
}

func (t *Test) Execute(ctx context.Context, arg args.Arguments) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("exit ping")
			return nil
		default:
			fmt.Println("ping")
			time.Sleep(time.Second * 1)
		}
	}
}
