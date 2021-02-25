### safecontainer包介绍
##### 三种实现方式
- 无锁队列实现 （ABA地址重用的问题没有解决，有隐患）
- Mutex加锁的实现方式（推荐使用）
- channel的实现方式（未完成）
	
##### 下面是测试数据
- 无锁队列的测试结果：
```go
	Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkList)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkList-8   	10938285	       105 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	github.com/giant-tech/go-service/base/safecontainer	1.480s
```

- 带mutex的测试结果
```go
	Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkMuList)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkMuList-8   	11244946	       104 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	github.com/giant-tech/go-service/base/safecontainer	1.498s
```
	- 单纯的channel测试结果
```go
Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkChan)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkChan-8   	17694469	        68.8 ns/op	       0 B/op	       0 allocs/op
PASS
```
