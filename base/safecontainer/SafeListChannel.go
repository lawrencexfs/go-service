package safecontainer

import "time"

//channel 版本的安全队列未完待续...

// SafeListChannelNode 节点
type SafeListChannelNode struct {
	Queue chan interface{}
}

// NewSafeListC 新建节点
func NewSafeListC(MaxQueueLen int) (sl *SafeListChannelNode) {
	sl = &SafeListChannelNode{}
	sl.Queue = make(chan interface{}, MaxQueueLen)
	return sl
}

// Push 非阻塞push
func (sl *SafeListChannelNode) Push(data interface{}, waittime time.Duration) bool {
	click := time.After(waittime)
	select {
	case sl.Queue <- data:
		return true
	case <-click:
		return false
	}
}

// Pop 非阻塞pop
// waittime Millisecond 一般10ms
func (sl *SafeListChannelNode) Pop(waittime time.Duration) (data interface{}) {
	click := time.After(waittime)
	select {
	case data = <-sl.Queue:
		return data
	case <-click:
		return nil
	}
}

// IsEmpty 是否为空
/*func (sl *SafeListChannelNode) IsEmpty() bool {
	//todo 判断channel是否为空
	return true
}
*/
