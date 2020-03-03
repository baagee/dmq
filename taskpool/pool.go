package taskpool

type Pool struct {
	TaskChannel  chan *Task
	workerNumber uint
	JobsChannel  chan *Task
}

//创建一个协程池
func NewPool(cap uint) *Pool {
	p := Pool{
		TaskChannel:  make(chan *Task),
		workerNumber: cap,
		JobsChannel:  make(chan *Task),
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
	for {
		task := <-p.TaskChannel
		p.JobsChannel <- task
	}
}

func (p *Pool) AddTask(task *Task) {
	p.TaskChannel <- task
}

func (p *Pool) Close() {
	close(p.JobsChannel)
	close(p.TaskChannel)
}
