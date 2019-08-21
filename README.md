# 学习Etcd-Raft

学习Etcd-Raft的demo

# Raft-Demo
## TODO
1. change to progress
2. add Debug module -> OK
3. lock for thread safe -> OK
4. election timeout
5. cpu occupancy -> OK
6. web front-end show -> WebAssembly
7. node break & rejoin
8. design log_unstable flush to log_stable
### 20190821
7. Follower node raf CRUD API: transfer to Leader to solve -> OK
8. script for test CRUD API
9. clear ReadIndex code -> OK

## FIXME
1. 前端日志会出现两次 -> Fixed 
原生Ajax需要判断状态
``` JavaScript
if (ajax.readyState !== 4 || ajax.status !== 200) {
            return;
}
```

2. CPU 占用过高的问题 -> Fixed
之前在main方法里面写了一个for死循环，导致CPU占用过高，每个进程占用CPU20%以上，五个进程开启之后CPU占用100%
因为里面有个HTTP SERVER，所以main方法里面去掉for循环，也不用在协程里面启动。
用pprof工具没有查出来问题