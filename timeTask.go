package timetask

import (
	"sync"
	"time"
)

// Task is a interface with method of DoTask
type Task interface {
	DoTask()
}

// TimeTask is a task with time to execute, with a task interface, and next timestamp
// to execute
type TimeTask struct {
	Task        Task
	NextTimeStp time.Time
	EveryTime   time.Duration
}

// NewTimeTask create a new timetask with task and duration
func NewTimeTask(t Task, d time.Duration) *TimeTask {
	nts := time.Now().Unix()
	nt := time.Unix(nts, 0)
	return &TimeTask{
		Task:        t,
		NextTimeStp: nt.Add(d),
		EveryTime:   d,
	}
}

func (t *TimeTask) nextTicker() *TimeTask {
	if t.EveryTime > 0 {
		t.NextTimeStp = t.NextTimeStp.Add(t.EveryTime)
	} else {
		return nil
	}

	return t
}

// TimeTaskList is a list of time tasks
type TimeTaskList struct {
	TaskList []*TimeTask
	TimeChan chan time.Time
	sync.Mutex
	WaitGroup sync.WaitGroup
	Quit      chan struct{}
}

// NewTimeTaskList create a new TimeTaskList
func NewTimeTaskList() *TimeTaskList {
	return &TimeTaskList{
		TaskList:  []*TimeTask{},
		TimeChan:  make(chan time.Time, 100),
		WaitGroup: sync.WaitGroup{},
		Quit:      make(chan struct{}),
	}
}

// Add timetask to TimeTaskList
func (l *TimeTaskList) Add(t *TimeTask) {
	l.Lock()
	defer l.Unlock()

	St := 0
	Ed := len(l.TaskList) - 1
	Len := Ed / 2

	if len(l.TaskList) == 0 {
		l.TaskList = append(l.TaskList, t)
		l.TimeChan <- t.NextTimeStp
		return
	}

	if l.TaskList[St].NextTimeStp.After(t.NextTimeStp) {
		tl := append([]*TimeTask{}, t)
		tl = append(tl, l.TaskList...)
		l.TaskList = tl
		l.TimeChan <- t.NextTimeStp
		return
	}

	if l.TaskList[Ed].NextTimeStp.Before(t.NextTimeStp) {
		l.TaskList = append(l.TaskList, t)
		return
	}

	for {
		if Ed-St <= 1 {
			Len = 1
			break
		}

		if l.TaskList[St+Len].NextTimeStp.After(t.NextTimeStp) {
			Ed = St + Len
			Len = (Ed - St) / 2
			continue
		}

		if l.TaskList[St+Len].NextTimeStp.Before(t.NextTimeStp) {
			St += Len
			Len = (Ed - St) / 2
			continue
		}

		if l.TaskList[St+Len].NextTimeStp.Equal(t.NextTimeStp) {
			break
		}
	}

	pos := St + Len
	tl := append([]*TimeTask{}, l.TaskList[pos:]...)
	l.TaskList = append(l.TaskList[:pos], t)
	l.TaskList = append(l.TaskList, tl...)
	return
}

// Parser time task list and do tasks
func (l *TimeTaskList) Parser() {
	var tc <-chan time.Time

	for {
		select {
		case t := <-l.TimeChan:
			if t.Before(time.Now()) {
				continue
			}

			tc = time.After(t.Sub(time.Now()))
		case <-tc:
			for {
				l.Lock()
				t := l.TaskList[0]

				if t.NextTimeStp.Before(time.Now()) || t.NextTimeStp.Equal(time.Now()) {
					l.WaitGroup.Add(1)
					go func(t *TimeTask) {
						defer l.WaitGroup.Done()
						t.Task.DoTask()
					}(t)
					l.TaskList = l.TaskList[1:]
					l.Unlock()

					nt := t.nextTicker()
					if nt != nil {
						l.Add(nt)
					}
				} else {
					l.Unlock()
					tc = time.After(t.NextTimeStp.Sub(time.Now()))
					break
				}
			}
		case <-l.Quit:
			l.WaitGroup.Wait()
			return
		}
	}
}
