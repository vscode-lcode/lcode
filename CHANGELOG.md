# Changelog

## [2.0.1] - 2023-01-13

### Change

- 更换 sqlite 驱动为 `github.com/mattn/go-sqlite3`, `modernc.org/sqlite`跨平台编译时出现错误
- 删除 `dd` 的 `status=none` 选项, 兼容 `busybox dd`
- 将 `client` 表移动到内存中了, 免去每次启动时都要清空表
- 服务端`lcode`命令现已支持命令行选项了, 输入 `--help` 可显示版本号
