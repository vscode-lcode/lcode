## 简介

webdav server over bash

一个基于 `bash` 的 `webdav server`, 文件的列出用的是 `ls`, 读取写入用的 `dd`, 所以只要有这三个就可以运行了

#### 用途/目标

使用本地编辑器编辑服务器文件

## 使用/安装

#### 用法 (在服务器上)

直接调用

```sh
>/dev/tcp/127.0.0.1/4349 0> >(echo 0) 0>&1  2> >(grep -E ^lo: >&2) bash -i -s
```

设置`alias`

```sh
# 服务器写入别名, 方便调用
echo "alias lcode='>/dev/tcp/127.0.0.1/4349 0> >(echo 0) 0>&1  2> >(grep -E ^lo: >&2) bash -i -s'" >> ~/.bashrc
source ~/.bashrc
```

通过`lcode`调用

```sh
# 打开webdav目录
lcode
# 输出编辑器链接, 点击打开编辑器进行编辑
lo: vscode://lcode.hub/shy-drone-f0_f0_f0_f0_f0_f0/root
```

debug

```sh
# 只输出执行的命令
>/dev/tcp/127.0.0.1/4349 0> >(echo 0) 0>&1  2> >(grep -vE '^\[' >&2) bash -i -s
# 输出所有 stderr
>/dev/tcp/127.0.0.1/4349 0> >(echo 0) 0>&1  2> >(cat >&2) bash -i -s
```

## 安装/设置 (本机)

### 下载 (暂无)

```sh
wget -O lcode-hub https://xxxxxx/ && chmod +x lcode-hub && sudo mv lcode-hub /usr/local/bin/
```

### 从源码 build

```sh
make build
# the binanry
./lcode-hub
# output
lcode-hub is running on 127.0.0.1:4349
```

### 设置 ssh config

```conf
# ~/.ssh/config
# config for lcode
Host *
  # 转发 hub 端口
  RemoteForward 127.0.0.1:4349 127.0.0.1:4349
  # 避免多次端口转发
  ControlMaster auto
  ControlPath /tmp/ssh_control_socket_%lcodeh_%p_%r
  # 启动lcode-hub
  LocalCommand lcode-hub &
  PermitLocalCommand yes
```

### 进阶设置/设计/FAQ

#### 是如何确定服务器的唯一性?

服务器的 ID 由两部分组成, `namespace`+`no`

- `namespace` 的获取来源是 `/proc/sys/kernel/hostname`
- `no` 是当前相同`namespace`下的服务器数量+1,
  但是服务器可以宣称自己的`no`是其他数字, 通过 `~/.lcode-id` 设置

这个服务器数量是保存在本地的`sqlite`数据库里`host`表中的, 路径是`~/.config/lcode/lcode.db`

注: 当服务器的 `~/.lcode-id` 不存在或值为`0`时, 会重新生成一个`no`并保存在服务器上

#### 是如何区分同一台服务器上的多个`lcode`

这是在 `client` 表中维护的, 每个`lcode`连接和`workdir`都会记录在该表中

如果请求的路径在该表中不存在, 那么就会返回 403 错误

## 一些开发中用到的技巧

#### 将 tcp socket 用作管道

```sh
echo 0 | 4>&0 5>/dev/tcp/127.0.0.1/4349 3> >(>&5 cat <&4) cat <&5 | cat
```
