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

## 使用示例

### 1. 自动保存上一条命令

```bash
$ ls -la /home/user/documents  # 执行一个普通命令
$ cs                          # 保存上一条命令
已保存命令: ls -la /home/user/documents
```

### 2. 手动保存命令

```bash
$ cs -y "docker run -d -p 80:80 nginx"  # 直接保存指定命令
已保存命令: docker run -d -p 80:80 nginx
```

### 3. 查看保存的命令

```bash
$ cs -l
ID  | 命令                              | 时间
1   | ls -la /home/user/documents       | 2023-12-23 14:30:25
2   | docker run -d -p 80:80 nginx      | 2023-12-23 14:35:10
```

### 4. 按天查看命令历史

```bash
$ cs -d
=== 2023-12-23 ===
- ls -la /home/user/documents
- docker run -d -p 80:80 nginx

=== 2023-12-22 ===
- git push origin main
- npm install express
```

### 5. 删除命令记录

```bash
$ cs -rm 1                    # 删除ID为1的命令记录
已删除ID为1的命令记录
```

### 6. 清理数据库

```bash
$ cs -c                       # 清理数据库
数据库已清理完成
```

## 开发

### 依赖

- Go 1.21.5 或更高版本
- SQLite3

### 构建

```bash
go build -o cs
```

## 许可证

MIT License 