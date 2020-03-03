package taskpool

import (
	"github.com/baagee/dmq/common"
)

type Pool struct {
	TaskChannel  chan *Task
	workerNumber uint
	JobsChannel  chan *Task
}

//创建一个协程池
func NewPool(size uint) *Pool {
	cc := common.Config.MsgDetailChanLen / 2
	p := Pool{
		TaskChannel:  make(chan *Task, cc),
		workerNumber: size,
		JobsChannel:  make(chan *Task, cc),
	}
	return &p
}

func (p *Pool) worker(workerId uint) {
	//worker不断的从JobsChannel内部任务队列中拿任务
	for task := range p.JobsChannel {
		task.Execute(workerId)
	}
}

func (p *Pool) Run() {
	var i uint
	for i = 0; i < p.workerNumber; i++ {
		go p.worker(i)
	}
	cc := common.Config.MsgDetailChanLen / 2
	for {
		if uint(len(p.JobsChannel)) < cc {
			p.JobsChannel <- <-p.TaskChannel
		}
	}
}

func (p *Pool) AddTask(task *Task) {
	cc := common.Config.MsgDetailChanLen / 2
	for {
		if uint(len(p.TaskChannel)) < cc {
			p.TaskChannel <- task
			break
		}
	}
}

func (p *Pool) Close() {
	close(p.JobsChannel)
	close(p.TaskChannel)
}
