package syncUtil

import (
	"sync"
)

// 多等1的实现
type MultiWaitOne struct {
	// 当前是否正在处理中
	isDoing bool

	// 处理中的锁对象
	doingLock sync.Mutex

	// 所有等待方
	allWaitChan []chan bool
}

// 开始执行或等待执行完成，只有一个能成功返回true
// 返回值:
// bool:是否开始成功
func (this *MultiWaitOne) StartDoOrWait() bool {
	myWaitChan := make(chan bool, 1)

	func() {
		this.doingLock.Lock()
		defer this.doingLock.Unlock()
		if this.isDoing == false {
			this.isDoing = true
			myWaitChan <- true
			return
		}

		// 加入到等待队列
		this.allWaitChan = append(this.allWaitChan, myWaitChan)
	}()

	return <-myWaitChan
}

// 完成处理
// 返回值:
// bool:是否是本次完成触发处理完成
func (this *MultiWaitOne) Done() bool {
	if this.isDoing == false {
		return false
	}

	this.doingLock.Lock()
	defer this.doingLock.Unlock()
	if this.isDoing == false {
		return false
	}

	// 标记为已处理完成
	this.isDoing = false

	if len(this.allWaitChan) <= 0 {
		return true
	}

	// 通知所有人，处理完成
	for _, item := range this.allWaitChan {
		item <- false
	}
	this.allWaitChan = make([]chan bool, 10)

	return true
}
