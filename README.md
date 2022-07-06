# 简介

起因是某位朋友抱怨 vim 太难用了, 想在服务器上使用本地的 vscode 进行编辑,
但 vscode remote ssh 太吃内存了, 简单编辑用 webdav 文件协议编辑也许可行?

也许会加上命令行支持?

# 进度

- [x] webdav
- [ ] edit with local vscode

# 设计概述

```sh
# open vscode for opening lcode plugin, lcode plugin will start server
code
ssh -R 4349:127.0.0.1:4349 root@your_host
# defualt connect server addr is 127.0.0.1:4349
lcode -c 127.0.0.1:4349
# log vscode link, you can click the link to open vscode
# if vscode lcode plugin is active will auto open
-> vscode://lcode-plugin/uuid-key
```
