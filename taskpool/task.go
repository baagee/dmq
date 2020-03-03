package taskpool

import "github.com/baagee/dmq/common"

/* 有关Task任务相关定义及操作 */
type Task struct {
	f func(workId uint) error
}

//通过NewTask来创建一个Task
func NewTask(f func(workId uint) error) *Task {
	t := Task{
		f: f,
	}
	return &t
}

//执行Task任务的方法
func (t *Task) Execute(workId uint) {
	err := t.f(workId)
	if err != nil {
		common.RecordError(err)
	}
}
