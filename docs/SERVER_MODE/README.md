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

Server data is stored in `./data` relative to the working directory. Use a dedicated directory, and keep that directory stable across restarts. It does not automatically share or synchronize the desktop database. Native desktop dialogs and window controls are not browser capabilities.

服务端数据保存在启动工作目录下的 `./data`；请使用独立目录，并在重启时保持相同工作目录。它不会自动共享或同步桌面版数据库；原生文件对话框、窗口控制也不是浏览器功能。

The example binds to loopback. Network deployment needs controlled access to the entire reader and API; this server does not provide a user login. See the [API reference](swagger.json).

以上示例仅监听本机。向网络开放时应对整个阅读器和 API 配置访问控制；服务端没有用户登录功能。接口见 [API 文档](swagger.json)。
