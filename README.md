## 主机编译与启动
在Linux端结束旧程序并编译和运行新程序, 复制这段到终端运行:

cd ~/agent_noah &&
go build -o ~/agent-bin/agent-noah.new ./agent/cmd &&
mv ~/agent-bin/agent-noah.new ~/agent-bin/agent-noah &&
~/agent-bin/agent-noah

比手写方便.

## 运行前准备：
1.本地创建.env.local文件，就是.env.example那个位置

## 运行：
终端输入: go run ./agent/cmd

## 须知:
如果你在本机跑需要注意 --当前各工具绑定"当前所在目录", 逻辑位于: agent/internal/workspace/workspace.go, 详情:"const DefaultDir = `.`".
正常来说没什么问题, 运行go run时cd到 '你的盘:\agent_noah>' 就行.

## 重要须知:
如果我的api_key意外泄露, 麻烦你联系 15166143379@163.com 或 he2619522@gmail.com.
Should my API key be accidentally leaked, please contact 15166143379@163.com or he2619522@gmail.com.

请不要默默消耗! 除非你是一个巨大的同性恋.
Don't silently consume! if you're a big gay.

## 附言：
记忆功能完成后不再隐藏提示词, 现在未达预期.

个人练习项目，项目目标为'协作者'.

近期做长任务需要的各组件,同时随便优化现有的东西.

第一版确认完成之前不写完整的README。

-狰和
2026/10/13
