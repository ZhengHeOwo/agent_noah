## 程序状态:

无法运行 / 升级架构途中顺便学习MySQL , 删除旧代码导致程序无法编译

## 开发计划:

预计10/17-19 完成'数据库与运行器'重构

## 主机编译与启动:

在Linux端结束旧程序并编译和运行新程序, 复制这段到终端运行:

cd ~/agent_noah &&\
go build -o ~/agent-bin/agent-noah.new ./agent/cmd &&\
mv ~/agent-bin/agent-noah.new ~/agent-bin/agent-noah &&\
~/agent-bin/agent-noah

比手写方便

## 运行前准备：

在'.env.example'相同位置, 按该文件提示创建'.env.local'文件

## 运行：

终端输入: go run ./agent/cmd

## 须知:

工具绑定"当前所在目录", 逻辑位于: agent/internal/workspace/workspace.go, 详情:"const DefaultDir = `.`"

正常来说没什么问题, 运行go run时cd到 '你的盘:\agent_noah>' 即可

程序的brave(网络搜索)是外国的, 建议使用时挂着VPN

## 附言：

提示词本身未达预期, 需随项目发展完善\

个人求职练习项目.

-ZhengHe\
2026/10/18
