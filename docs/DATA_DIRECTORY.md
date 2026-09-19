# Custom data directory / 自定义数据目录

## Desktop settings / 桌面设置

Open **Settings → General → Data Management → Data directory**, browse to an
**existing empty folder**, and choose **Change directory**. Fully quit MrRSS,
including its tray icon, then reopen it. Before opening SQLite, MrRSS copies the
whole directory into a staging folder and switches only after the copy succeeds.
The original directory remains as a backup. Verify subscriptions, reading state,
favorites, credentials and custom scripts before removing that backup yourself.
You can cancel a pending change in settings. Failure leaves the original directory
active and shows an error in settings. Symbolic links and special files require
manual migration. Do not launch another version or server against these folders
during migration.

打开 **设置 → 常规 → 数据管理 → 数据目录**，浏览并选择一个**现有空文件夹**，
点击“更改目录”。完全退出（包括系统托盘）后重新打开，程序会在打开 SQLite 前复制
完整数据，复制成功后切换。原目录保留作为备份；核对订阅、阅读状态、收藏、凭据和
自定义脚本后再自行清理。可在设置中取消待生效变更。失败时继续使用原目录，并在
设置中提示原因。包含符号链接或特殊文件的目录需要手动迁移。迁移期间不要用其他
版本或服务端打开相关目录。

The bootstrap location is `MrRSS-bootstrap/storage.json` under the OS user config
directory, or `mrrss-storage.json` beside a portable executable. It is separate from
the selected database so ordinary app launches and startup integration remember
the location. This setting is not synchronized between computers.

启动位置配置保存在系统用户配置目录下的 `MrRSS-bootstrap/storage.json`，便携版则
保存在可执行文件旁的 `mrrss-storage.json`。它独立于数据库，普通启动和开机启动均
会使用该位置；此设置不跨设备同步。

## Advanced launch options / 高级启动选项

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
Keep the option in manual shortcuts too. Removing it restores the directory chosen in desktop settings (or the default).
Launch overrides do not migrate existing data automatically.

内置开机启动会保留自定义路径；手动创建的快捷方式也应保留参数。去掉参数和环境变量后恢复使用软件内设置的目录（未设置时为默认目录）。
启动参数本身不会自动迁移数据。

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
