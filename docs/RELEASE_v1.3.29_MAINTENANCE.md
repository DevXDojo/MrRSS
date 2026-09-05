# v1.3.29 维护分组

盘点日期：2026-09-05。工作分支：`release/v1.3.29`。

执行约定：每个功能或修复单独提交；整组完成后统一验证；每完成一组完全停止，收到用户继续指令后才处理下一组。分组是候选队列，不承诺一次解决所有条目。遇到需要大改或缺少复现资料的条目记录原因并暂缓，不扩大本轮范围。不自动发布版本、不合并 main、不自动关闭缺乏验证证据的 issue。

当前轮次：第一、二组已完成，本轮结束后停止。已审查并整合 PR #1047、#1048、#1049、#1050、#1051、#1052、#1053，修复后纳入发布分支。#1054 是 release/v1.3.29 → main 的发布草稿，保持不合并。

排序原则：先处理有明确复现和现成修复的高影响故障，再处理局部交互、管理功能及正文问题；平台依赖和缺资料问题先确认复现。组内大致按影响与实现成本排序。云同步、移动端及跨模块大改暂缓。

## 第一组：现有 PR 与高影响故障（已完成）

- [x] [#1046](https://github.com/DevXDojo/MrRSS/issues/1046) [BUG] Auto-refresh runs in an endless loop (thousands of times/second) when update interval is set to a large value (e.g. 46080 minutes)
- [x] [#1044](https://github.com/DevXDojo/MrRSS/issues/1044) [BUG] AI translation gets 502 from OpenAI-compatible gateway while profile test passes
- [x] [#1043](https://github.com/DevXDojo/MrRSS/issues/1043) [BUG] 通过软件更新失败
- [x] [#874](https://github.com/DevXDojo/MrRSS/issues/874) [BUG] 刷新信息列表的时候，页面会闪，用户体验不好，另外右边的正文页内容会变空白
- [x] [#909](https://github.com/DevXDojo/MrRSS/issues/909) [BUG] 我仍然不能全文翻译

## 第二组：小范围阅读交互（已完成）

- [x] [#696](https://github.com/DevXDojo/MrRSS/issues/696) [STYLE] 按钮悬停提示时有时无
- [x] [#567](https://github.com/DevXDojo/MrRSS/issues/567) [FEATURE] 悬浮提示中显示快捷键
- [x] [#561](https://github.com/DevXDojo/MrRSS/issues/561) [FEATURE] 支持快速返回原订阅源
- [x] [#427](https://github.com/DevXDojo/MrRSS/issues/427) [FEATURE] 右上角工具栏添加复制链接按钮
- [x] [#358](https://github.com/DevXDojo/MrRSS/issues/358) [FEATURE] 支持右键搜索选中文本
- [x] [#736](https://github.com/DevXDojo/MrRSS/issues/736) [FEATURE] 关于翻译功能的建议
- [x] [#779](https://github.com/DevXDojo/MrRSS/issues/779) [FEATURE] B站视频查看模式

## 第三组：分类与订阅管理（待继续）

- [ ] [#757](https://github.com/DevXDojo/MrRSS/issues/757) [FEATURE] 分类文件夹可以拖动排序，类似对订阅源进行的拖动排序操作
- [ ] [#587](https://github.com/DevXDojo/MrRSS/issues/587) [FEATURE] 可调整订阅源分组上下位置（可与订阅拖动调整方式相同）
- [ ] [#501](https://github.com/DevXDojo/MrRSS/issues/501) [BUG] 拖动订阅源时蓝色的框跳动
- [ ] [#653](https://github.com/DevXDojo/MrRSS/issues/653) [FEATURE] 收藏夹改进建议
- [ ] [#455](https://github.com/DevXDojo/MrRSS/issues/455) [FEATURE] 自动化规则的导出与导入功能
- [ ] [#508](https://github.com/DevXDojo/MrRSS/issues/508) [FEATURE] 侧边栏分类功能增强
- [ ] [#548](https://github.com/DevXDojo/MrRSS/issues/548) [FEATURE] 建议增加置顶和多种配排序方式

## 第四组：正文、抓取与渲染（先复现再实现，待继续）

- [ ] [#795](https://github.com/DevXDojo/MrRSS/issues/795) [BUG] 有时rss订阅只剩标题了
- [ ] [#805](https://github.com/DevXDojo/MrRSS/issues/805) [BUG] 自建FOLO服务器订阅连接的订阅内容自动消失
- [ ] [#948](https://github.com/DevXDojo/MrRSS/issues/948) 部分文章无法正常渲染
- [ ] [#799](https://github.com/DevXDojo/MrRSS/issues/799) [FEATURE] 文中内容涉及markdown内容的渲染问题
- [ ] [#605](https://github.com/DevXDojo/MrRSS/issues/605) [BUG] 图片模式，图片太多页面下滑，布局会崩掉
- [ ] [#982](https://github.com/DevXDojo/MrRSS/issues/982) [FEATURE] papr有一个自动正文的功能很方便看文，能不能加入类似的功能
- [ ] [#601](https://github.com/DevXDojo/MrRSS/issues/601) [FEATURE] 获取全文能力有待提高
- [ ] [#908](https://github.com/DevXDojo/MrRSS/issues/908) [FEATURE] 希望考虑添加freshrss类似的全文css选择器
- [ ] [#828](https://github.com/DevXDojo/MrRSS/issues/828) [FEATURE] 希望增加一个cookies选项，来替换失效的cookies

## 第五组：平台与安装故障（需要对应环境，待继续）

- [ ] [#777](https://github.com/DevXDojo/MrRSS/issues/777) [FEATURE] 建议程序运行状态下，安装可以直接关闭程序
- [ ] [#320](https://github.com/DevXDojo/MrRSS/issues/320) 按关闭时最小化到托盘后，再打开时软件窗口不是最大化，而变成中间一块，我是绿色版，2560*1600， 125%缩放
- [ ] [#661](https://github.com/DevXDojo/MrRSS/issues/661) [STYLE] 窗口边缘判定触发窗口缩放的斜箭头
- [ ] [#426](https://github.com/DevXDojo/MrRSS/issues/426) [BUG] 右侧窗口缩放手势被错误识别为滚动条操作
- [ ] [#447](https://github.com/DevXDojo/MrRSS/issues/447) [BUG] 订阅高级设置页面存在明显的 UI 抖动
- [ ] [#796](https://github.com/DevXDojo/MrRSS/issues/796) [BUG] 更新后无法关闭及打开app「附视频」
- [ ] [#556](https://github.com/DevXDojo/MrRSS/issues/556) [BUG] 一直在报错修改Mac上的app
- [ ] [#852](https://github.com/DevXDojo/MrRSS/issues/852) [BUG] arch linux 下运行白屏
- [ ] [#771](https://github.com/DevXDojo/MrRSS/issues/771) [BUG] linux debian 13版本运行帧率低
- [ ] [#626](https://github.com/DevXDojo/MrRSS/issues/626) [BUG] 软件布局样式暂未生效，预计 1 分钟后恢复正常
- [ ] [#699](https://github.com/DevXDojo/MrRSS/issues/699) winget 防病毒产品报告安装程序受感染
- [ ] [#800](https://github.com/DevXDojo/MrRSS/issues/800) [BUG] 重装系统后数据库损坏
- [ ] [#544](https://github.com/DevXDojo/MrRSS/issues/544) [BUG] Cannot display unread message badge in the status bar

## 第六组：已支持能力核对及缺资料事项（待继续）

- [ ] [#918](https://github.com/DevXDojo/MrRSS/issues/918) [BUG] 不准备维护了还是？翻译功能根本无法使用，直接翻译错误，换了腾讯云翻译api一样的保持，无语！
- [ ] [#916](https://github.com/DevXDojo/MrRSS/issues/916) rdf订阅添加进去没有文章，请问可以解决吗
- [ ] [#825](https://github.com/DevXDojo/MrRSS/issues/825) [BUG] 香港的claw中国优化服务器，搭建的freshrss
- [ ] [#772](https://github.com/DevXDojo/MrRSS/issues/772) [BUG] AI摘要功能 无法使用
- [ ] [#767](https://github.com/DevXDojo/MrRSS/issues/767) [BUG] 翻译和ai问题
- [ ] [#335](https://github.com/DevXDojo/MrRSS/issues/335) [BUG] 部分订阅无法显示图标
- [ ] [#672](https://github.com/DevXDojo/MrRSS/issues/672) [FEATURE] 希望可以增加用户自定义CSS外观的功能
- [ ] [#421](https://github.com/DevXDojo/MrRSS/issues/421) [FEATURE] 能否提供在浏览器中运行的网页版界面
- [ ] [#542](https://github.com/DevXDojo/MrRSS/issues/542) [FEATURE] 建议支持模型增加gemini
- [ ] [#941](https://github.com/DevXDojo/MrRSS/issues/941) [FEATURE] 能不能整合一些opml 配置文件
- [ ] [#543](https://github.com/DevXDojo/MrRSS/issues/543) 【讨论】有订阅华尔街日报的方法么，官方rss无法订阅
- [ ] [#546](https://github.com/DevXDojo/MrRSS/issues/546) [FEATURE] 希望可以分类排序、红书的视频封面预览、视频类
- [ ] [#435](https://github.com/DevXDojo/MrRSS/issues/435) [BUG] 快速点击订阅源时的视觉闪烁

- [ ] [#656](https://github.com/DevXDojo/MrRSS/issues/656) [FEATURE] AI api增加claude支持（已有 Anthropic 协议实现，后续核对界面及端到端行为。）

## 暂缓：大改、长期项目或发布策略

- [ ] [#91](https://github.com/DevXDojo/MrRSS/issues/91) Support cloud syncing / 支持云同步
- [ ] [#92](https://github.com/DevXDojo/MrRSS/issues/92) Support for mobile devices / 支持移动端
- [ ] [#981](https://github.com/DevXDojo/MrRSS/issues/981) 能否出一个UWP版本并且支持动态磁贴的
- [ ] [#945](https://github.com/DevXDojo/MrRSS/issues/945) Change view...
- [ ] [#943](https://github.com/DevXDojo/MrRSS/issues/943) [FEATURE] 是否考虑支持mactype渲染？
- [ ] [#914](https://github.com/DevXDojo/MrRSS/issues/914) [FEATURE] 视频、推送等服务
- [ ] [#832](https://github.com/DevXDojo/MrRSS/issues/832) [FEATURE] 建议增加一个最小化到任务栏后，桌面放一个滚动条，像DesktopTicker那种
- [ ] [#829](https://github.com/DevXDojo/MrRSS/issues/829) [FEATURE] 支持对指定数量文章生成汇总报告
- [ ] [#688](https://github.com/DevXDojo/MrRSS/issues/688) [FEATURE] 能否支持思源笔记集成服务
- [ ] [#673](https://github.com/DevXDojo/MrRSS/issues/673) [FEATURE] 希望增加自定义缓存路径
- [ ] [#660](https://github.com/DevXDojo/MrRSS/issues/660) [FEATURE] 突出显示功能，高亮条目或文本
- [ ] [#658](https://github.com/DevXDojo/MrRSS/issues/658) [FEATURE] 日报/汇报/时间线功能
- [ ] [#630](https://github.com/DevXDojo/MrRSS/issues/630) [FEATURE] 建议收费
- [ ] [#603](https://github.com/DevXDojo/MrRSS/issues/603) [BUG] xml文本内容太多读取失败
- [ ] [#580](https://github.com/DevXDojo/MrRSS/issues/580) [FEATURE] 稍后阅读的优化与体验改进
- [ ] [#577](https://github.com/DevXDojo/MrRSS/issues/577) [FEATURE] 桌面通知功能
- [ ] [#576](https://github.com/DevXDojo/MrRSS/issues/576) [FEATURE] 收藏夹单独管理和组织
- [ ] [#565](https://github.com/DevXDojo/MrRSS/issues/565) [FEATURE] 按日期或订阅源分组列表
- [ ] [#564](https://github.com/DevXDojo/MrRSS/issues/564) [FEATURE] 支持先预览再订阅
- [ ] [#563](https://github.com/DevXDojo/MrRSS/issues/563) [FEATURE] 日期相关改进建议
- [ ] [#468](https://github.com/DevXDojo/MrRSS/issues/468) [FEATURE] 显示原网页可以支持原生webview
- [ ] [#467](https://github.com/DevXDojo/MrRSS/issues/467) [BUG] 学术期刊RSS订阅无法显示原网页，完整文章内容获取不全
- [ ] [#400](https://github.com/DevXDojo/MrRSS/issues/400) [FEATURE] 基于相同URL或标题的去重功能
- [ ] [#316](https://github.com/DevXDojo/MrRSS/issues/316) [FEATURE] 成熟以后强烈建议上载微软商店
- [ ] [#303](https://github.com/DevXDojo/MrRSS/issues/303) [FEATURE] 希望可以完善 macOS 客户端
- [ ] [#104](https://github.com/DevXDojo/MrRSS/issues/104) Add auto-labeling feature with automatic generation on viewport entry
- [ ] [#90](https://github.com/DevXDojo/MrRSS/issues/90) Optimize article list with virtual scrolling

## 第一组审查与验证记录

第一组完成后已停止，后续继续指令启动第二组。上述勾选表示发布分支已有修复，不表示已发布或已关闭 GitHub issue。

| PR / issue | 发布分支整合提交 | 审查结论与补强 |
| --- | --- | --- |
| #1051 / #1046 | `d501b791` | 超长定时器分段等待；保留 fixed 模式限制；卸载时取消自动刷新。 |
| #1052 / #874 | `5b1eab9b` | 刷新保留所选文章；补充翻页去重、旧请求淘汰、失败时保留正文。 |
| #1050 / #1044 | `b50c16a2` | AI 测试及实际请求统一网络配置；补充代理凭据转义和 IPv6 地址处理。 |
| #1053 / #909 | `f9ac1178` | 图文混排正文翻译；补充行内链接/强调保留、代码/公式保护及旧翻译状态隔离。 |
| #1049 / #1043 | `7b0fb71f` | 实际字节进度、有限重试和续传；补充独立临时目录、严格范围响应检查、网络错误手动回退。 |
| #1047 | `7c3c6ac9` | 官网 Vue、ESLint、typescript-eslint 依赖更新。 |
| #1048 | `913dffdd` | goquery 与 Wails 更新；同步 `@wailsio/runtime` 到 beta.15。 |

组末统一验证：

- `go test -timeout=5m ./...`：通过。
- 前端 Vitest：91/91 通过，涵盖超长刷新、保留选中、翻页去重、导航竞态、图文段落处理。
- Cypress：更新流程 15/15、正文翻译 1/1，通过。使用临时目录内的隔离服务，不使用用户订阅数据库；没有执行真实安装。
- 前端 ESLint：0 errors；本轮改动的 src 文件以 `--max-warnings 0` 通过。全库仍有既有 CRLF/格式警告，不在此组扩大清理范围。
- 前端锁文件 `npm ci --dry-run --ignore-scripts`：通过；官网 `npm ci` 和生产构建：通过。
- `wails3 build`：通过，产物 `build/bin/MrRSS.exe`；另已通过隔离测试服务的 server 构建。本轮使用临时目录中的 Wails beta.15 工具，不替换系统原有 CLI。
- `git diff --check`：通过。更新接口文档及生成的技能 API 参考已同步，下载进度路由按实际 `/api/download-update/progress` 记录。
- `CHANGELOG.md`：已在 Unreleased 中逐项记录本组修复和依赖更新；最终发布时间确定前不虚构发布日期。

验证边界：本机为 Windows，未进行 macOS/Linux 原生桌面运行验证；Vite 仍提示既有大体积打包产物。旧版本机 Swagger 工具的依赖扫描无法解析 Go 1.27 标准库语法，本组使用内部类型扫描取得更新接口 schema，仅同步涉及下载的端点，避免混入既有文档漂移。

第一组后的下一组为 #696、#567、#561 等局部交互；详见后续第二组记录。发布草稿 #1054 保持不合并。

远端核对：2026-09-05，GitHub 已确认 #1047–#1053 全部为 MERGED，目标均为 `release/v1.3.29`。#1054 保持 OPEN / draft；main 仍为 `92c39341`。隔离浏览器回归使用的临时服务已停止。


## 第二组审查与验证记录

本组完成后完全停止，第三组及以后待用户明确继续。每项功能独立提交，统一在组末验证。

| Issue | 功能提交 | 实现范围 |
| --- | --- | --- |
| #696 | `ea276cf1` | 装饰图标不再抢占按钮悬停命中区域。 |
| #567 | `40e20eee` | 导航、文章、搜索和工具栏提示显示当前快捷键；自定义或禁用后同步更新。 |
| #561 | `203334ac` | 点击来源或右键返回订阅源；清除筛选和搜索并保留较早文章，关闭阅读弹窗，淘汰旧筛选结果。 |
| #427 | `56b9b219` | 工具栏已有复制入口；补充浏览器剪贴板、原生回退及失败提示。 |
| #358 | `b049dcf9` | 正文选中文字后可右键选择 Google、Bing、百度或 DuckDuckGo；选择搜索引擎后才传出文本。 |
| #736 | `162e95e4` | 新增按需翻译设置；标题按钮及 Ctrl/Command 点击单段；保留默认自动翻译，支持失败重试。 |
| #779 | `aca6ae85` | 图片/视频瀑布流遵循订阅级及全局外部浏览器偏好，显式正文模式可覆盖全局偏好。 |

本组同时补充剪贴板回退、跨筛选返回来源和旧请求隔离的单元测试，以及阅读交互浏览器回归。按需翻译不自动提交全文；原生输入框、链接及媒体右键菜单保留。云同步、移动端及其他大改未纳入本组。

组末统一验证结果：

- `go test -timeout=5m ./...`：通过。
- Vitest：95/95 通过，包括剪贴板回退、返回订阅源保留选中及旧筛选响应隔离。
- Cypress：22/22 通过（阅读交互 6、正文自动翻译 1、更新流程 15）；搜索与外部打开均截获请求，未实际向搜索服务传送文本。
- 本组改动 src 文件 ESLint：`--max-warnings 0` 通过。
- 前端生产构建和 `wails3 build`：通过，Windows 产物 `build/bin/MrRSS.exe`。使用与项目一致的 Wails beta.15 工具。
- `git diff --check`：通过。生成的设置文件经统一格式化后仅保留新设置的必要差异。
- 已将本组全部功能记录到 `CHANGELOG.md` 的 Unreleased；没有修改历史版本的功能记录。
- 额外导航补强提交 `b5b8c5e5`，统一验证与格式整理另行提交。构建产生的锁文件平台元数据变化已排除。

验证边界：本机验证 Windows；未运行 macOS/Linux 原生桌面测试。Vite 仍有既有的大包体提示。初次 Go 检查与前端打包发生嵌入资源替换竞争，打包完成后重跑已通过；浏览器回归初次出现的选择器误报已修正并完整重跑通过。

停点：第二组完成，隔离测试服务已停止。下一组为“分类与订阅管理”；本轮不启动第三组，不合并发布草稿 #1054，不发布版本。
