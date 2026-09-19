# Custom data directory / 自定义数据目录

Desktop and server builds accept `--data-dir PATH`, or the `MRRSS_DATA_DIR`
environment variable. The command line takes precedence. Relative paths resolve
against the launch working directory; use an absolute path in shortcuts. The
option overrides the portable/server/default directory and applies to the whole
data directory, including `rss.db`, logs, scripts and caches. An invalid or
unwritable directory stops startup instead of silently opening another database.

桌面版和服务端均支持 `--data-dir 路径` 或环境变量 `MRRSS_DATA_DIR`，命令行优先。
相对路径以启动工作目录为准，快捷方式建议使用绝对路径。此选项覆盖便携版、服务端和
普通安装的默认路径，并统一作用于数据库、日志、脚本和缓存。目录无法写入时会停止启动。

```powershell
# Windows shortcut target can use the same arguments.
& 'C:\Program Files\MrRSS\MrRSS.exe' --data-dir 'D:\My Reader Data'
```

```sh
./MrRSS.AppImage --data-dir "$HOME/Reader Data"
open -a MrRSS --args --data-dir "$HOME/Reader Data"
./mrrss-server --host 127.0.0.1 --data-dir /srv/mrrss-data
```

Built-in start-on-login registration preserves the resolved custom directory.
Keep the option in manual shortcuts too. Removing it selects the original
default directory again; no data is deleted or moved automatically.

内置开机启动会保留自定义路径；手动创建的快捷方式也应保留参数。去掉参数和环境变量后
恢复使用默认目录，程序不会自动删除或迁移数据。

## Move existing data / 迁移现有数据

1. Completely quit every MrRSS process using the original directory.
2. Back up and copy the **entire** original data directory into the destination.
   Keep all files, including SQLite sidecar files and scripts.
3. Start with `--data-dir` pointing to the copied directory and verify feeds,
   reading state, favorites and settings. Keep the backup until verified.

先完全退出使用原目录的所有 MrRSS 进程，备份并复制整个数据目录（含 SQLite 附属文件、
脚本等），再通过新路径启动，检查订阅、阅读状态、收藏和设置。确认前保留原备份。
空目录会创建独立资料库。

This selects storage; it does not synchronize databases or resolve cloud-drive
conflicts. Machine-bound encrypted credentials may need to be entered again on
another computer. 用户可自行管理云盘，但此选项只选择存储位置，不提供数据库同步或冲突
合并。换电脑时，绑定机器加密的凭据可能需要重新输入。
