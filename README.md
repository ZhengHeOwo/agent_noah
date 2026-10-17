## 主机编译与启动

在Linux端结束旧程序并编译和运行新程序, 复制这段到终端运行:

cd ~/agent_noah &&\
go build -o ~/agent-bin/agent-noah.new ./agent/cmd &&\
mv ~/agent-bin/agent-noah.new ~/agent-bin/agent-noah &&\
~/agent-bin/agent-noah

比手写方便

## 运行前准备：

1.本地创建.env.local文件，就是.env.example那个位置

## 运行：

终端输入: go run ./agent/cmd

## 须知:

如果你在本机跑需要注意 --当前各工具绑定"当前所在目录", 逻辑位于: agent/internal/workspace/workspace.go, 详情:"const DefaultDir = `.`"

正常来说没什么问题, 运行go run时cd到 '你的盘:\agent_noah>' 就行

程序的brave(网络搜索)是外国的, 建议使用时挂着VPN

如果我的api_key意外泄露, 麻烦你联系 15166143379@163.com 或 he2619522@gmail.com

Should my API key be accidentally leaked, please contact 15166143379@163.com or he2619522@gmail.com

## 附言：

提示词本身未达预期, 需随项目发展完善\

个人求职练习项目.

前缀污染暂无立即解决办法,或许可以尝试在前缀加入随机数函数作为指纹在短期阻断该行为,但模型是模仿,不确定模仿内容改变是否影响模仿行为,暂不处理,等后续架构重制引入Claude的设计思路.

10/17-19完成现有sqlite重构

第一版确认完成之前不写完整的README

-ZhengHe\
2026/10/17
