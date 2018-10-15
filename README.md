# zookeeper_practice
practice zookeeper with go-zookeeper

1. 高可用,主从机制, 主服务,从同步, 主挂掉时,自动选举主节点自动完成服务恢复
2. 服务发现,动态添加删除从节点

目标: 实现一个简易日志采集中心, 满足高可用强一致

zookeeper tree
------------------------------------------------------------------------
[zk: localhost:2181(CONNECTED) 4] ls /<br>
[myconn, zookeeper]<br>
[zk: localhost:2181(CONNECTED) 5] ls /myconn<br>
[process_list, master]<br>
[zk: localhost:2181(CONNECTED) 6] ls /myconn/process_list<br>
[001, 002]<br>
[zk: localhost:2181(CONNECTED) 7] get /myconn/master<br>
2202<br>
