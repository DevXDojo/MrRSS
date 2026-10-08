# Browser interface / 浏览器界面

The server build serves the same reader interface at `/` and data endpoints at `/api/*`. The desktop application's loopback API listener serves API endpoints only; opening that listener in a browser does not provide the reader UI.

server 构建在 `/` 提供阅读器界面，在 `/api/*` 提供数据接口。桌面程序的本机 API 监听器只提供接口，不提供浏览器阅读页面。

## Build and open locally / 本机构建与打开

Build the frontend before embedding it in the server. From the repository root:

先构建前端，再将其嵌入服务端。从仓库根目录执行：

```sh
cd frontend
npm ci
npm run build
cd ..
go build -tags server -o mrrss-server .
./mrrss-server --host 127.0.0.1 --port 1234
```

On Windows use `go build -tags server -o mrrss-server.exe .` and run `./mrrss-server.exe --host 127.0.0.1 --port 1234` in PowerShell.

Windows 下将输出命名为 `mrrss-server.exe`，在 PowerShell 中使用 `./mrrss-server.exe --host 127.0.0.1 --port 1234` 启动。

Open `http://127.0.0.1:1234/` in a browser. Change the port if the desktop app is also running and using 1234. Browser translation can operate on the rendered article text; availability depends on the browser and its network access.

浏览器打开 `http://127.0.0.1:1234/`。桌面版同时运行并占用 1234 时，请换一个端口。浏览器翻译可作用于渲染的文章文字，能否使用取决于浏览器及其网络连接。

Server data defaults to `./data` relative to the working directory. Use `--data-dir PATH` or `MRRSS_DATA_DIR` to choose another location; see [custom data directories](../DATA_DIRECTORY.md). Keep the directory stable across restarts. It does not automatically share or synchronize the desktop database. Native desktop dialogs and window controls are not browser capabilities.

服务端数据默认保存在启动工作目录下的 `./data`，可通过 `--data-dir 路径` 或 `MRRSS_DATA_DIR` 指定位置，详见[自定义数据目录](../DATA_DIRECTORY.md)。重启时请保持目录一致。它不会自动共享或同步桌面版数据库；原生文件对话框、窗口控制也不是浏览器功能。

The example binds to loopback. Network deployment needs controlled access to the entire reader and API; this server does not provide a user login. See the [API reference](swagger.json).

以上示例仅监听本机。向网络开放时应对整个阅读器和 API 配置访问控制；服务端没有用户登录功能。接口见 [API 文档](swagger.json)。

## GHCR image access / 容器镜像访问

Public repository visibility does not make the GHCR package public. If an anonymous pull of `ghcr.io/devxdojo/mrrss:<version>` returns 403, a package administrator must open [mrrss Package settings](https://github.com/orgs/DevXDojo/packages/container/mrrss/settings) and change the package visibility to **Public**. If the organization restricts public packages, an organization owner must allow this first. This is a registry setting, not an application code change. See [GitHub's access and visibility documentation](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility).

仓库公开不代表 GHCR 包公开。匿名拉取 `ghcr.io/devxdojo/mrrss:<版本>` 返回 403 时，需要包管理员打开 [mrrss 包设置](https://github.com/orgs/DevXDojo/packages/container/mrrss/settings)，将包的可见性改为 **Public**。组织限制公开包时，需组织所有者先允许公开。这是镜像仓库权限设置，无法仅靠修改应用代码解决。详见 [GitHub 权限文档](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility)。

The Docker release workflow checks the published version's manifest using an empty Docker configuration. A successful authenticated push is no longer enough for a successful release job. After changing visibility, verify with an anonymous pull (use an empty `DOCKER_CONFIG` to avoid cached credentials):

Docker 发布流程会使用空 Docker 配置检查版本镜像的匿名访问；登录后推送成功不再代表发布检查通过。修改可见性后，用未登录的 Docker 配置确认可拉取镜像：

```sh
docker pull ghcr.io/devxdojo/mrrss:1.3.40
```
