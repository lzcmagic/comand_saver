# Command Saver (cs)

一个简单的命令行工具，用于保存和管理你的命令行历史记录。

## 功能特点

- 自动保存上一条执行的命令
- 支持手动添加命令和描述
- 按时间顺序查看保存的命令
- 按天查看最近7天的命令历史
- 支持删除指定的命令记录
- 支持数据库清理

## 安装

从 [Releases](https://github.com/YOUR_USERNAME/command_saver/releases) 页面下载适合你系统的最新版本，解压后将可执行文件 `cs` 放到系统的 PATH 目录中。

## 使用方法

```bash
cs                  # 保存上一条执行的命令
cs -l               # 列出所有保存的命令
cs -d               # 按天显示最近7天的命令
cs -y <命令>        # 直接保存指定的命令
cs -rm <id>         # 删除指定ID的命令记录
cs -h               # 显示帮助信息
cs -c               # 清理数据库
```

## 开发

### 依赖

- Go 1.23.4 或更高版本
- SQLite3

### 构建

```bash
go build -o cs
```

## 许可证

MIT License 