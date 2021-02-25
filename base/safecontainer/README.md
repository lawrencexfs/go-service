### safecontainer������
##### ����ʵ�ַ�ʽ
- ��������ʵ�� ��ABA��ַ���õ�����û�н������������
- Mutex������ʵ�ַ�ʽ���Ƽ�ʹ�ã�
- channel��ʵ�ַ�ʽ��δ��ɣ�
	
##### �����ǲ�������
- �������еĲ��Խ����
```go
	Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkList)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkList-8   	10938285	       105 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	github.com/giant-tech/go-service/base/safecontainer	1.480s
```

- ��mutex�Ĳ��Խ��
```go
	Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkMuList)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkMuList-8   	11244946	       104 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	github.com/giant-tech/go-service/base/safecontainer	1.498s
```
	- ������channel���Խ��
```go
Running tool: D:\Go\bin\go.exe test -benchmem -run=^$ -bench ^(BenchmarkChan)$ github.com/giant-tech/go-service/base/safecontainer

goos: windows
goarch: amd64
pkg: github.com/giant-tech/go-service/base/safecontainer
BenchmarkChan-8   	17694469	        68.8 ns/op	       0 B/op	       0 allocs/op
PASS
```
