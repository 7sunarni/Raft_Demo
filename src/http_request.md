# Raft中用来请求的api

**1. 读的api，用于读取数据**
``` shell
curl 'http://localhost:9001/raft' -X POST --data '{"Operation":"READ","Key":"key"}'
```
**2. 写的api，用于存放数据**
``` shell
curl 'http://localhost:9001/raft' -X POST --data '{"Operation":"ADD","Key":"apdo","Value":100}'
```

**3. debug的api，用于读取日志信息**
``` shell

```