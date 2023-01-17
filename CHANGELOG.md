# Changelog

## [2.1.8] - 2023-01-17

### Fix

- 为 net.Conn 设置超时时间, 避免连接一直挂起, 从而导致编辑器一直挂起
- 修复编辑目标是文件夹是以/结尾时无法通过路径结尾不带/的路径访问 (因为 vscode 第一次访问是路径末尾不带/访问)

## [2.1.7] - 2023-01-17

### Fix

- 修复 namespace 含大写字母时无法访问的问题. 原因是因为域名会自动转小写导致无法找到对应的 host, 所以将 id 统一小写化

## [2.1.6] - 2023-01-17

- 将 hid 添加到命令参数中, 方便与其他编辑器集成(指[vscode lcode hub](https://github.com/vscode-lcode/hub))

## [2.1.5] - 2023-01-16

### Fix

- webdav 访问域名添加随机 hub-id 避免服务器通过网络权限探测&获取权限外的文件
  > lcode-id 在服务器上, namespace 也在服务器上, 所以服务器的 webdav 访问域名很容易就可以从服务器中生成, 如果服务器上有恶意程序定时获取特定文件而你刚好打开该文件目录的访问权限的话, 恶意程序便可通过网络访问权限获取到它本不该/不可获取的文件, 这次修复加上随机 hub-id 让服务器无法生成 webdav 访问地址避免了这种问题

## [2.1.4] - 2023-01-15

### Change

- 更换 sqlite 驱动为 `modernc.org/sqlite`, `github.com/mattn/go-sqlite3`跨平台编译时出现错误
- 添加了多平台编译配置

## [2.1.2] - 2023-01-15

### Add

- 添加 github workflow release

## [2.1.1] - 2023-01-15

### Add

- `lcode-hub` 添加日志等级设置, 默认日志等级设置为 `Info`

## [2.1.0] - 2023-01-13

### Add

- `lcode` 支持打开多个编辑目标了

## [2.0.1] - 2023-01-13

### Change

- 更换 sqlite 驱动为 `github.com/mattn/go-sqlite3`, `modernc.org/sqlite`跨平台编译时出现错误
- 删除 `dd` 的 `status=none` 选项, 兼容 `busybox dd`
- 将 `client` 表移动到内存中了, 免去每次启动时都要清空表
- 服务端`lcode`命令现已支持命令行选项了, 输入 `--help` 可显示版本号
