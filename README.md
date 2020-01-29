# MulticenterABEForFabric
高校区块链-密码服务平台比赛专用修改TMACS



启动RPC server，默认启动（9999、10000、10001、10002）

```shell
./startserver.sh
```

关闭全部RPC server

```
./stopserver.sh
```

单独关闭相应的端口进程，port要写入具体的端口号

```shell
kill -9 `lsof -t -i:$port`
```

