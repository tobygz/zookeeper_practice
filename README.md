# zookeeper_practice
practice zookeeper with go-zookeeper

1. 高可用,主从机制, 主服务,从同步, 主挂掉时,自动选举主节点自动完成服务恢复
2. 服务发现,动态添加删除从节点

目标: 实现一个简易日志采集中心, 满足高可用强一致

zookeeper tree
------------------------------------------------------------------------
[zk: localhost:2181(CONNECTED) 4] ls /
[myconn, zookeeper]
[zk: localhost:2181(CONNECTED) 5] ls /myconn
[process_list, master]
[zk: localhost:2181(CONNECTED) 6] ls /myconn/process_list
[001, 002]
[zk: localhost:2181(CONNECTED) 7] get /myconn/master
2202
