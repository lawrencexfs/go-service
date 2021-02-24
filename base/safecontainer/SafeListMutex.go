package safecontainer

import "sync"

// SafeListMutexNode 节点
type SafeListMutexNode struct {
	next  *SafeListMutexNode
	value interface{}
}

// newNodeM 新建节点
func newNodeM(data interface{}) *SafeListMutexNode {
	return &SafeListMutexNode{next: nil, value: data}
}

// SafeListMutex 安全链表
type SafeListMutex struct {
	head *SafeListMutexNode
	tail *SafeListMutexNode

	mu sync.Mutex
	// 有数据就用channel通知
	HasDataC chan bool
}

// NewSafeListM 新创建一个列表
func NewSafeListM() *SafeListMutex {
	return &SafeListMutex{
		HasDataC: make(chan bool, 1),
	}
}

// Put 放入
func (sl *SafeListMutex) Put(data interface{}) {
	sl.mu.Lock()
	newNode := newNodeM(data)

	if sl.tail != nil {
		sl.tail.next = newNode
		sl.tail = newNode
	} else {
		sl.tail = newNode
		sl.head = newNode
	}
	sl.mu.Unlock()
	select {
	case sl.HasDataC <- true:
	default:
	}

}

// Pop 拿出
func (sl *SafeListMutex) Pop() (interface{}, error) {

	sl.mu.Lock()
	defer sl.mu.Unlock()
	if sl.tail == nil {
		return nil, errNoNode
	}

	if sl.head == sl.tail {
		v := sl.head
		sl.head = nil
		sl.tail = nil
		return v.value, nil
	}

	v := sl.head
	sl.head = sl.head.next
	return v.value, nil
}

// IsEmpty 是否为空
func (sl *SafeListMutex) IsEmpty() bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	ret := (sl.tail == nil)
	return ret
}
