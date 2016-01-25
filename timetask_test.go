package timetask

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
# File Name: test_timeTask.go
# Author: lianming
# mail: lianming@alibaba-inc.com
# Created Time: 2016年01月25日 星期一 10时06分18秒
*/

type PrtTask struct {
	Str string
}

func (t *PrtTask) DoTask() {
	fmt.Printf("[%s] %s\n", time.Now(), t.Str)
}

func Test_tickertask(t *testing.T) {
	wg := sync.WaitGroup{}

	l := NewTimeTaskList()

	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Parser()
	}()

	t1 := &PrtTask{"1 Sec Task"}
	t2 := &PrtTask{"2 Sec Task"}
	t3 := &PrtTask{"3 Sec Task"}

	tt1 := NewTimeTask(t1, time.Second)
	tt2 := NewTimeTask(t2, time.Second*2)
	tt3 := NewTimeTask(t3, time.Second*3)

	l.Add(tt1)
	l.Add(tt2)
	l.Add(tt3)

	wg.Wait()
	t.Logf("Exit")
}
