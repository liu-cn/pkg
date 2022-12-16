package benchmark

import (
	"fmt"
	"sync"
	"time"
)

func Time(fn func(), runNum int) Info {
	now := time.Now()
	for i := 0; i < runNum; i++ {
		fn()
	}
	return Info{
		start: now,
		end:   time.Now(),
	}
}
func TimeAndRes(fn func() interface{}, runNum int) Info {
	now := time.Now()
	var res interface{}
	for i := 0; i < runNum; i++ {
		v := fn()
		if i == runNum-1 {
			res = v
		}
	}
	return Info{
		start:   now,
		end:     time.Now(),
		LastRes: res,
	}
}

type Info struct {
	start   time.Time
	end     time.Time
	LastRes interface{}
}

func (i *Info) Print(str ...interface{}) {
	s := ""
	for _, v := range str {
		s += fmt.Sprintf("%v", v)
	}

	fmt.Println(s, i.end.Sub(i.start))
}

func TimeSync(fn func(), runNum int) Info {
	i := Info{
		start: time.Now(),
	}
	wg := sync.WaitGroup{}
	wg.Add(runNum)
	for i := 0; i < runNum; i++ {
		go func() {
			fn()
			wg.Done()
		}()
	}
	wg.Wait()
	i.end = time.Now()
	return i
}
