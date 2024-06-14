$paths = @(
    "tmp/my-raft-cluster",
    "tmp/my-raft-cluster/node1",
    "tmp/my-raft-cluster/node2",
    "tmp/my-raft-cluster/node3"
)

foreach ($path in $paths) {
    # 创建目录，如果目录已存在，则不做任何操作
    New-Item -Path $path -ItemType Directory -Force
}

go build
#
# $commands = @(
#     @{Command="./neochain.exe --raft_bootstrap --raft_id=node1 --address=localhost:50051 --raft_data_dir tmp/my-raft-cluster"; LogFile="node1.log"; Name="neochain-node1"},
#     @{Command="./neochain.exe --raft_id=node2 --address=localhost:50052 --raft_data_dir tmp/my-raft-cluster"; LogFile="node2.log"; Name="neochain-node2"},
#     @{Command="./neochain.exe --raft_id=node3 --address=localhost:50053 --raft_data_dir tmp/my-raft-cluster"; LogFile="node3.log"; Name="neochain-node3"}
# )
#
# foreach ($cmd in $commands) {
#     $job = Start-Job -ScriptBlock {
#         param($command, $logFile)
#         Set-Location $workingDirectory
#         & cmd.exe /c $command > $logFile
#     } -ArgumentList $cmd.Command, $cmd.LogFile -Name $cmd.Name
# }
#
# # 显示所有作业的状态
# Get-Job | Format-Table Id, State, HasMoreData, Command, Location
#
# go install github.com/Jille/raftadmin/cmd/raftadmin@latest
#
# # raftadmin localhost:50051 add_voter node2 localhost:50052 0 &
# # raftadmin --leader multi:///localhost:50051,localhost:50052 add_voter node3 localhost:50053 0 &
