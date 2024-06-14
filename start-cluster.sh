#!/bin/bash

rm -r tmp
rm neochain

# 创建目录
mkdir -p tmp/my-raft-cluster/node{A,B,C}

# 编译项目
go build

killall neochain

# 启动三个raft节点
# 启动三个raft节点，端口号分别为50051, 50052, 50053
./neochain --raft_bootstrap --raft_id=nodeA --address=localhost:50051 --raft_data_dir tmp/my-raft-cluster > tmp/my-raft-cluster/nodeA/system.log 2>&1  &
disown
./neochain --raft_id=nodeB --address=localhost:50052 --raft_data_dir tmp/my-raft-cluster > tmp/my-raft-cluster/nodeB/system.log 2>&1  &
disown
./neochain --raft_id=nodeC --address=localhost:50053 --raft_data_dir tmp/my-raft-cluster > tmp/my-raft-cluster/nodeC/system.log 2>&1  &
disown

# 安装raftadmin
go install github.com/Jille/raftadmin/cmd/raftadmin@latest

# 等待节点启动
echo "waiting for the cluster to start"
sleep 5

# 配置raft节点
raftadmin localhost:50051 add_voter nodeB localhost:50052 0
raftadmin --leader multi:///localhost:50051,localhost:50052 add_voter nodeC localhost:50053 0

# 等待用户按回车继续
read -p "Press enter to continue"

# 运行hammer.go
go run hammer/hammer.go

# 转移领导权
#raftadmin --leader multi:///localhost:50051,localhost:50052,localhost:50053 leadership_transfer

# 等待所有后台进程完成
