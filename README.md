# 基于 raft 共识协议和 LSM 架构引擎的 key-value 储存服务系统

实现了 mit 6.824 的 lab 2、3 并通过所有测试
同时使用了 LSM 架构的数据库系统来为 lab3 提供的数据储存服务

![屏幕截图 2023-09-22 145222.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/fdd84964a3ec40878936d2f2a582bc87~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=907&h=1024&s=163004&e=png&b=1e2030)
![屏幕截图 2023-09-22 145204.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/6f80626c2c864691981b5ac1a599dc06~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=901&h=1063&s=195305&e=png&b=1e2030)

ps:<br> 1.系统为了实现多种极端情况下的测试，暂定没有实现 RPC,任然使用的 labrpc,这样可以更方便的测试,如果要实现部署使用,可以使用 gRPC 等成熟的 RPC 解决方案实现;<br>2.暂时只支持 string:string 的储存.
