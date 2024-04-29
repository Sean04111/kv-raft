# 基于 raft 共识协议和 LSM 架构引擎的 key-value 储存服务系统

## 项目架构

项目在架构上使用了基于 mit6.824 的 raft 协议实现，使用 raft 协议来实现项目的共识，使用了 lsm tree 来实现数据储存服务。
![image.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/11e093fd22d7461398199ca3c467d06f~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=1825&h=1200&s=118639&e=png&b=fdf8f6)

## 项目模块

### Raft

项目的 raft 模块使用了基于 mit 6.824 lab2 的 raft 实现思路，具体的实现参照 raft-extended 论文实现；
实现了 raft 协议中的 leader 选举，日志复制、日志对齐、日志压缩快照等功能，实现了基本的共识协商；

> 不足之处：
> 这里没有实现对于集群中成员变更的情况下的共识的实现
> 统一使用的是粒度比较大的 mutex，可以考虑使用粒度更小的 atomic

### KV service&clerk

基于 lab3 实现了一个基本的 kv 客户端和服务端的交互，
实现了基本的 client 端和 server 端之间的交互，同时保证了命令的顺序一致性和强一致性；

> 不足之处:
> 为了方便地实现顺序一致性，让读请求也只能通过 leader 来实现，这样效率有点低，可以划分事务请求的方式，让非写请求在所有节点都可以访问，而写请求只能在 leader 节点访问，这样效率会更高；

### lsm tree

基于 lsm Tree 的思想为系统实现了数据存储功能，实现了内存表的读写，预写日志防止内存表丢失，内存表向磁盘 sstable 转移，以及 sstable 在磁盘中的分层管理；

源码流程图：
![image.png](https://p1-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/5f0ad4921f744dffafbf74cdc0922c41~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=2377&h=1004&s=166882&e=png&b=fdf8f6)

后面这里用了跳表来优化,(没有选择红黑树是因为实现太复杂了,性能收益不大),从benchmark来看,从读取的性能能优化25%左右：<br>
<br>基于BST的:
![image.png](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/5eed481552ba48649bf53ea033ce4846~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=1304&h=360&s=50683&e=png&b=1e2030)
<br>基于跳表的：
![image.png](https://p9-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/f1628da535194d73ba84b6733611db1c~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=1354&h=493&s=76879&e=png&b=1e2030)

> 不足之处:
>跳表的锁粒度太大了（这里还是使用的读写锁），可以换成左边界锁来优化

ps:<br> 1.系统为了实现多种极端情况下的测试，暂定没有实现 RPC,任然使用的 labrpc,这样可以更方便的测试,如果要实现部署使用,可以使用 gRPC 等成熟的 RPC 解决方案实现;<br>2.暂时只支持 string:string 的储存.

## 项目测试

项目已经通过所有 mit 测试,代码覆盖率都在 90%以上

![屏幕截图 2023-09-22 145222.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/fdd84964a3ec40878936d2f2a582bc87~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=907&h=1024&s=163004&e=png&b=1e2030)
![屏幕截图 2023-09-22 145204.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/6f80626c2c864691981b5ac1a599dc06~tplv-k3u1fbpfcp-jj-mark:0:0:0:0:q75.image#?w=901&h=1063&s=195305&e=png&b=1e2030)
