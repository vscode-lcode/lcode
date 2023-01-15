# Changelog

## [2.1.3] - 2023-01-15

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
