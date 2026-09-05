# v1.3.29 维护分组

盘点日期：2026-09-05。工作分支：`release/v1.3.29`。

执行约定：每个功能或修复单独提交；整组完成后统一验证；每完成一组完全停止，收到用户继续指令后才处理下一组。分组是候选队列，不承诺一次解决所有条目。遇到需要大改或缺少复现资料的条目记录原因并暂缓，不扩大本轮范围。不自动发布版本、不合并 main、不自动关闭缺乏验证证据的 issue。

当前轮次：前四组已完成；用户明确授权第五、六组合并处理。本轮 27 个条目已审查，7 项完成代码/文档，20 项已按 issue 语言回复后暂缓；组末统一验证后停止。已审查并整合 PR #1047、#1048、#1049、#1050、#1051、#1052、#1053，修复后纳入发布分支。#1054 是 release/v1.3.29 → main 的发布草稿，保持不合并。

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

## 第三组：分类与订阅管理（已完成）

- [x] [#757](https://github.com/DevXDojo/MrRSS/issues/757) [FEATURE] 分类文件夹可以拖动排序，类似对订阅源进行的拖动排序操作
- [x] [#587](https://github.com/DevXDojo/MrRSS/issues/587) [FEATURE] 可调整订阅源分组上下位置（可与订阅拖动调整方式相同）
- [x] [#501](https://github.com/DevXDojo/MrRSS/issues/501) [BUG] 拖动订阅源时蓝色的框跳动
- [x] [#653](https://github.com/DevXDojo/MrRSS/issues/653) [FEATURE] 收藏夹改进建议
- [x] [#455](https://github.com/DevXDojo/MrRSS/issues/455) [FEATURE] 自动化规则的导出与导入功能
- [x] [#508](https://github.com/DevXDojo/MrRSS/issues/508) [FEATURE] 侧边栏分类功能增强
- [x] [#548](https://github.com/DevXDojo/MrRSS/issues/548) [FEATURE] 建议增加置顶和多种配排序方式

## 第四组：正文、抓取与渲染（已完成）

- [x] [#795](https://github.com/DevXDojo/MrRSS/issues/795) [BUG] 有时rss订阅只剩标题了
- [x] [#805](https://github.com/DevXDojo/MrRSS/issues/805) [BUG] 自建FOLO服务器订阅连接的订阅内容自动消失
- [x] [#948](https://github.com/DevXDojo/MrRSS/issues/948) 部分文章无法正常渲染
- [x] [#799](https://github.com/DevXDojo/MrRSS/issues/799) [FEATURE] 文中内容涉及markdown内容的渲染问题
- [x] [#605](https://github.com/DevXDojo/MrRSS/issues/605) [BUG] 图片模式，图片太多页面下滑，布局会崩掉
- [x] [#982](https://github.com/DevXDojo/MrRSS/issues/982) [FEATURE] papr有一个自动正文的功能很方便看文，能不能加入类似的功能
- [x] [#601](https://github.com/DevXDojo/MrRSS/issues/601) [FEATURE] 获取全文能力有待提高
- [x] [#908](https://github.com/DevXDojo/MrRSS/issues/908) [FEATURE] 希望考虑添加freshrss类似的全文css选择器
- [x] [#828](https://github.com/DevXDojo/MrRSS/issues/828) [FEATURE] 希望增加一个cookies选项，来替换失效的cookies

## 第五组：平台与安装故障（已处理，含明确暂缓项）

- [ ] [#777](https://github.com/DevXDojo/MrRSS/issues/777) [FEATURE] 建议程序运行状态下，安装可以直接关闭程序（已回复，暂缓）
- [x] [#320](https://github.com/DevXDojo/MrRSS/issues/320) 按关闭时最小化到托盘后，再打开时软件窗口不是最大化，而变成中间一块，我是绿色版，2560*1600， 125%缩放
- [ ] [#661](https://github.com/DevXDojo/MrRSS/issues/661) [STYLE] 窗口边缘判定触发窗口缩放的斜箭头（已回复，暂缓）
- [ ] [#426](https://github.com/DevXDojo/MrRSS/issues/426) [BUG] 右侧窗口缩放手势被错误识别为滚动条操作（已回复，暂缓）
- [ ] [#447](https://github.com/DevXDojo/MrRSS/issues/447) [BUG] 订阅高级设置页面存在明显的 UI 抖动（已回复，暂缓）
- [x] [#796](https://github.com/DevXDojo/MrRSS/issues/796) [BUG] 更新后无法关闭及打开app「附视频」
- [ ] [#556](https://github.com/DevXDojo/MrRSS/issues/556) [BUG] 一直在报错修改Mac上的app（已回复，暂缓）
- [ ] [#852](https://github.com/DevXDojo/MrRSS/issues/852) [BUG] arch linux 下运行白屏（已回复，暂缓）
- [ ] [#771](https://github.com/DevXDojo/MrRSS/issues/771) [BUG] linux debian 13版本运行帧率低（已回复，暂缓）
- [x] [#626](https://github.com/DevXDojo/MrRSS/issues/626) [BUG] 软件布局样式暂未生效，预计 1 分钟后恢复正常
- [ ] [#699](https://github.com/DevXDojo/MrRSS/issues/699) winget 防病毒产品报告安装程序受感染（已回复，暂缓）
- [ ] [#800](https://github.com/DevXDojo/MrRSS/issues/800) [BUG] 重装系统后数据库损坏（已回复，暂缓）
- [ ] [#544](https://github.com/DevXDojo/MrRSS/issues/544) [BUG] Cannot display unread message badge in the status bar（已回复，暂缓）

## 第六组：已支持能力核对及缺资料事项（已处理，含明确暂缓项）

- [ ] [#918](https://github.com/DevXDojo/MrRSS/issues/918) [BUG] 不准备维护了还是？翻译功能根本无法使用，直接翻译错误，换了腾讯云翻译api一样的保持，无语！（已回复，暂缓）
- [ ] [#916](https://github.com/DevXDojo/MrRSS/issues/916) rdf订阅添加进去没有文章，请问可以解决吗（已回复，暂缓）
- [ ] [#825](https://github.com/DevXDojo/MrRSS/issues/825) [BUG] 香港的claw中国优化服务器，搭建的freshrss（已回复，暂缓）
- [ ] [#772](https://github.com/DevXDojo/MrRSS/issues/772) [BUG] AI摘要功能 无法使用（已回复，暂缓）
- [ ] [#767](https://github.com/DevXDojo/MrRSS/issues/767) [BUG] 翻译和ai问题（已回复，暂缓）
- [ ] [#335](https://github.com/DevXDojo/MrRSS/issues/335) [BUG] 部分订阅无法显示图标（已回复，暂缓）
- [x] [#672](https://github.com/DevXDojo/MrRSS/issues/672) [FEATURE] 希望可以增加用户自定义CSS外观的功能
- [x] [#421](https://github.com/DevXDojo/MrRSS/issues/421) [FEATURE] 能否提供在浏览器中运行的网页版界面
- [x] [#542](https://github.com/DevXDojo/MrRSS/issues/542) [FEATURE] 建议支持模型增加gemini
- [ ] [#941](https://github.com/DevXDojo/MrRSS/issues/941) [FEATURE] 能不能整合一些opml 配置文件（已回复，暂缓）
- [ ] [#543](https://github.com/DevXDojo/MrRSS/issues/543) 【讨论】有订阅华尔街日报的方法么，官方rss无法订阅（已回复，暂缓）
- [ ] [#546](https://github.com/DevXDojo/MrRSS/issues/546) [FEATURE] 希望可以分类排序、红书的视频封面预览、视频类（已回复，暂缓）
- [ ] [#435](https://github.com/DevXDojo/MrRSS/issues/435) [BUG] 快速点击订阅源时的视觉闪烁（已回复，暂缓）

- [x] [#656](https://github.com/DevXDojo/MrRSS/issues/656) [FEATURE] AI api增加claude支持（已有 Anthropic 协议实现，后续核对界面及端到端行为。）

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


## 第三组审查与验证记录

本组完成后停止，第四组及以后需用户明确继续。新增约定：每组完成后更新 PR #1054 的 description，每条已完成 issue 独占一行 `Fixed #编号`；并核对 GitHub 的 closingIssuesReferences。该发布草稿仍保持 release/v1.3.29 → main，暂不合并。

| Issue | 功能提交 | 实现范围 |
| --- | --- | --- |
| #653 | `49ac36a9` | 收藏计数使用全部收藏数量；汇总嵌套分类；已读收藏不受“仅未读”偏好影响。 |
| #501 | `d064cfd7` | 图标及间隙使用稳定行目标；插入线不占布局空间；中点防抖及拖动自动滚动；补齐嵌套事件传递。 |
| #757、#587 | `b8b039db` | 同级分类可拖动排序，支持嵌套分类且保存后重启保留，不改变层级或订阅归属。 |
| #508 | `ebc62c4b` | 分类右键可解散或取消整个分类的订阅；明确提示子分类及收藏影响；事务失败回滚；保护 FreshRSS 订阅。 |
| #455 | `0b5ef61c` | 导出 MrRSS 版本化 JSON；校验导入格式、字段和动作；追加且重新分配 ID，保留启用状态；不立即处理旧文章。 |
| #548 | `6bfed17c` | 同级分类/订阅置顶；名称升降序、数量升降序、最新文章、自定义顺序；选项持久化，拖动需自定义模式。 |

组末验证：

- `go test -timeout=5m ./...`：通过；覆盖分类事务回滚、字面分类前缀、保留/删除文章、FreshRSS 保护、接口参数及方法验证、已读收藏行为。
- Vitest：109/109 通过；新增规则备份往返、格式拒绝、ID 隔离及排序优先级测试。
- 本轮改动的 src 文件 ESLint：`--max-warnings 0` 通过。
- 浏览器回归包含分类根节点及嵌套拖动、图标拖放、排序和置顶持久化、收藏计数、批量操作确认、规则导入/导出，并覆盖前两组阅读与更新流程。
- 新 `/api/feeds/category` 接口已同步 Swagger 与由其生成的 API 参考。changelog 已逐项更新。

注意：规则备份使用 MrRSS 自有版本化格式，也接受旧设置中的原始规则数组，不宣称兼容其他阅读器的备份格式。解散操作将该分类及子分类的订阅移入“未分类”并保留文章；取消订阅会删除其文章，执行前明确确认。

最终核对：浏览器回归 28/28 通过（本组 6、阅读交互 6、正文翻译 1、更新 15）；`wails3 build` 通过，Windows 产物为 `build/bin/MrRSS.exe`。测试服务已停止，未操作用户订阅数据库。构建生成的锁文件平台元数据噪声已排除，`git diff --check` 通过。

PR #1054 已补齐三组共 19 条逐行 `Fixed #编号`，GitHub 的 closingIssuesReferences 已识别全部 19 个 issue；目标为默认分支 main、状态 OPEN / draft。待该发布 PR 正式合并时自动关闭关联 issue，本轮不合并它。

停点：第三组已完成，第四组“正文、抓取与渲染”尚未开始。总体维护目标仍有后续分组，需用户明确继续。

## 第四组审查与验证记录

用户已明确继续第四组，并允许对不便复现的外部站点问题依据代码修复。组内九个条目分别提交；统一在组末验证，完成本组后完全停止，第五组不自动启动。

| Issue | 功能提交 | 实现范围 |
| --- | --- | --- |
| #795 | `e05f1833` | 重载时清理数据库和内存两层缓存；空缓存不阻止重试；自动按年龄清理保留收藏和稍后读正文。 |
| #805 | `815af13d` | 移除每批保存时启动的清理任务；整轮刷新完成后单个清理任务按最旧内容小批处理；保留一天内新缓存和收藏、稍后读内容，允许保护内容暂时超过容量目标。 |
| #948 | `9ecdc4ab` | RSS 窗口之外的旧文章可从原网页恢复；失败时回退数据库保存的源摘要；单篇查询补齐 original_summary；阅读请求使用代次与取消控制。 |
| #982 | `f28c8902` | 自动全文对空正文同样生效；每次选择只自动请求一次，手动可重试；文章切换、重载和卸载时清理请求状态。 |
| #799 | `a8fa0a3b` | 识别纯 Markdown 订阅正文并渲染标题、表格、围栏代码；保留语言标记、数学内容与内嵌栅格图片，统一整理正文 HTML。 |
| #601 | `28cc46c7` | 处理页面字符编码、重定向后的相对地址与常见延迟图片；自动提取失败时尝试语义正文容器；全文响应与文章绑定，避免过期结果覆盖。 |
| #908 | `4ce69751` | 独立保存每个订阅源的正文和移除 CSS 选择器；多处匹配按顺序组合且嵌套去重；无匹配明确失败；FreshRSS 订阅也能通过右键配置本机提取规则。 |
| #828 | `e81dc7fd` | 每订阅源可替换或删除 Cookie；加密存储、界面只显示保存状态；按明确的网站 origin 限定发送，重定向到其他站点不携带凭据。 |
| #605 | `4def619e` | 使用已解码图片比例平衡瀑布流，延迟图片预留空间，批量更新布局并保留可见卡片位置；处理窄列、重复页、过期请求和监听器清理。 |

验证使用本机隔离数据库、HTTP 测试站点和浏览器拦截样例，没有修改用户订阅或真实网站 Cookie。原 issue 未提供可重现链接的场景，依据代码缺陷与可控回归样例验证；不宣称已实测所有报告中的私人 RSSHub/FOLO 服务、微信文章或微博视频。Cookie 只影响 MrRSS 发出的订阅/文章请求，不修改远端 RSSHub 或 FreshRSS 服务的认证配置。

组末检查：

- 后端全量测试：`go test -timeout=5m ./...`。新增缓存重载/恢复、保留策略、选择器提取、延迟图片、响应失败与取消、Cookie 加密/更新/删除/重定向隔离及接口验证。刷新集成测试改用临时文件数据库并释放资源，避免多连接 `:memory:` 的独立数据库误报。
- 前端单元测试：115/115；覆盖自动全文、失败重试、A→B→A 过期结果隔离、卸载取消、实际图片比例布局、窄容器与观察器初始化。
- 浏览器回归：33/33（本组 5、阅读 6、分类管理 6、正文翻译 1、更新 15）；包含 60 张延迟图片翻页、缩放和布局检查。
- 改动的 src 文件 ESLint `--max-warnings 0`；前端生产构建和 Windows `wails3 build`。
- 更新 CHANGELOG、Swagger 及生成的 API 参考；新增 `/api/feeds/content-options`。全局设置 schema 没有新增项，提取规则保存在以 feed_id 关联的独立订阅配置表中。

PR #1054 已包含前四组共 28 条逐行 `Fixed #编号`；GitHub closingIssuesReferences 已识别全部 28 个 issue。发布草稿继续保持 OPEN / draft；本组不合并 main，不发布版本。隔离测试服务在回归结束后关闭。

停点：第四组已完成，第五组“平台与安装故障”尚未开始。总体维护目标仍有后续分组，等待用户明确继续。

最终核对：后端全量测试、115 项前端单元测试、33 项浏览器回归、改动 src 文件的 ESLint 和 Windows `wails3 build` 全部通过；工作区排除构建产生的锁文件平台元数据改动。隔离测试服务已停止；main 仍为 `92c39341`。


## 第五、六组联合审查与验证记录

本轮遵照用户明确指令合并执行第五、六组，并获准按原 issue 语言回复资料不足或需大改的条目后跳过。27 项均已审查，其中 7 项完成，20 项回复后暂缓；暂缓项保持开放，不写入 PR 自动关闭清单。第七组长期/大改事项继续按原约定排除。

| Issue | 功能提交 | 本轮交付 |
| --- | --- | --- |
| #626 | `de644651` | 移除启动页面同步加载的外部字体及图标脚本；使用本地/系统字体和已打包 Vue 图标。 |
| #672 | `24cf9d32` | 将自定义 CSS 移到应用生命周期；无正文时生效，替换/删除具有取消及代次保护，更新中英文说明。 |
| #320 | `825955f1` | 统一托盘/菜单/二次启动恢复；不重放窗口尺寸与位置，不通过 Restore 取消最大化。 |
| #796 | `ffe9d1e6` | 移除 macOS 500ms 双击关闭要求；全屏退出完成后隐藏，重新打开取消待隐藏并聚焦。 |
| #656 | `1e18d179` | 补全 Claude 原生 Messages 端点识别、基本地址规范化、聊天 system 字段与错误保留。 |
| #542 | `44ddda88` | 区分 Gemini 原生/兼容协议及认证；使用选定模型生成原生路径，保留多段正文与系统指令。 |
| #421 | `7963f8d0` | 核对已有 server 浏览器界面；新增前端/服务端构建、本机访问、数据目录及功能边界中英文说明。 |

Claude 与 Gemini 原生协议在同一 AI 客户端中供测试、摘要、翻译和聊天调用。本轮使用拦截 HTTP transport 验证多轮请求、认证、端点、响应和失败分类，没有使用用户 API Key，也未调用收费服务。原生协议说明参考 [Claude Messages](https://platform.claude.com/docs/en/api/messages/create)、[Gemini generateContent](https://ai.google.dev/api/generate-content) 和 [Gemini 兼容协议](https://ai.google.dev/gemini-api/docs/openai)。

| 暂缓 issue / 已发布回复 | 原因 |
| --- | --- |
| [#777](https://github.com/DevXDojo/MrRSS/issues/777#issuecomment-5551543191) | 需要安装器与旧版本进程正常退出协调，暂缓安装生命周期改动。 |
| [#661](https://github.com/DevXDojo/MrRSS/issues/661#issuecomment-5551543366) | 缺系统/DPI/窗口状态，暂缓原生命中区改动。 |
| [#426](https://github.com/DevXDojo/MrRSS/issues/426#issuecomment-5551543549) | 缺原生滚动条/缩放边界复现环境。 |
| [#447](https://github.com/DevXDojo/MrRSS/issues/447#issuecomment-5551543714) | 原先首项已消失；剩余 1px 位移缺稳定条件。 |
| [#556](https://github.com/DevXDojo/MrRSS/issues/556#issuecomment-5551543881) | 缺弹窗文字、安装/启动/更新阶段与硬件信息。 |
| [#852](https://github.com/DevXDojo/MrRSS/issues/852#issuecomment-5551544022) | GBM 创建失败，缺显卡/驱动/WebKitGTK 组合。 |
| [#771](https://github.com/DevXDojo/MrRSS/issues/771#issuecomment-5551544174) | 缺桌面环境、图形栈和低帧率操作范围。 |
| [#699](https://github.com/DevXDojo/MrRSS/issues/699#issuecomment-5551544317) | 旧 winget PR 已关闭；缺检测名称、哈希及安全情报版本。 |
| [#800](https://github.com/DevXDojo/MrRSS/issues/800#issuecomment-5551544464) | 代码 14 是打开文件失败；更正删除 WAL 的旧建议，需路径和权限诊断。 |
| [#544](https://github.com/DevXDojo/MrRSS/issues/544#issuecomment-5551544582) | 现框架已有 macOS tray label，但计数接入与原生验证未实现。 |
| [#918](https://github.com/DevXDojo/MrRSS/issues/918#issuecomment-5551544742) | 腾讯云等失败不能从通用 AI 修复推定解决，缺具体服务与错误。 |
| [#916](https://github.com/DevXDojo/MrRSS/issues/916#issuecomment-5551544884) | 解析库已有 RDF，缺订阅地址或最小 XML。 |
| [#825](https://github.com/DevXDojo/MrRSS/issues/825#issuecomment-5551545036) | 拉取覆盖本地待同步记录风险需联同持久队列、重试、并发设计修复。 |
| [#772](https://github.com/DevXDojo/MrRSS/issues/772#issuecomment-5551545183) | 特定套餐/长摘要失败缺状态码和最小样例。 |
| [#767](https://github.com/DevXDojo/MrRSS/issues/767#issuecomment-5551545347) | Gemini 测试与默认 Google 翻译是不同路径，需分别诊断。 |
| [#335](https://github.com/DevXDojo/MrRSS/issues/335#issuecomment-5551545504) | 重启已恢复图标，追加的规则问题缺样例。 |
| [#941](https://github.com/DevXDojo/MrRSS/issues/941#issuecomment-5551545645) | 推荐 OPML/排行榜涉及列表来源与持续内容维护。 |
| [#543](https://github.com/DevXDojo/MrRSS/issues/543#issuecomment-5551545783) | 检查环境直连 HTTP 401，未取得实际 XML，需脱敏源样例。 |
| [#546](https://github.com/DevXDojo/MrRSS/issues/546#issuecomment-5551545922) | 分类排序已实现，视频封面和视频类需求缺样例；混合 issue 保留。 |
| [#435](https://github.com/DevXDojo/MrRSS/issues/435#issuecomment-5551546049) | 旧请求隔离已有修复，视觉无闪烁仍需列表与导航过渡调整。 |

组末验证：后端全量测试、119 项前端单元测试、改动 src 文件 ESLint 和 Windows `wails3 build` 已通过。新增 AI 原生/兼容协议回归、已识别协议失败单次请求与安全错误分类、原生窗口恢复调用序列、CSS 应用生命周期与请求竞态回归。隔离 server 构建及浏览器首页 HTTP 200 已确认；浏览器回归 35/35 全部通过（本轮 2、正文 5、阅读 6、分类管理 6、正文翻译 1、更新 15）。

验证边界：本机是 Windows；macOS 关闭/全屏逻辑依据代码及 Wails beta.15 原生事件实现修复，尚未进行 macOS 原生桌面交互测试，Linux 图形问题未做原生复现。启动外部资源阻塞已从代码移除，不宣称复现了报告者的公司网络。无用户数据或真实 API 凭据参与测试。

最终核对：GitHub 当前 82 个开放 issue 全部在清单中，无遗漏或重复；共 35 项已完成，20 项本轮已回复暂缓，27 项原定大改/长期事项继续排除。PR #1054 已更新 35 条逐行 `Fixed #编号`，closingIssuesReferences 与这 35 条完全一致，状态 OPEN / draft、release/v1.3.29 → main。前端锁文件仅构建平台元数据变化，已还原；隔离测试服务已停止。

停点：第五、六组联合处理完成，本轮完全停止。当前授权范围内的六个分组均已处理；暂缓项目不代表已修复，不自动推进原定排除的大改、不合并发布 PR、不发布版本。
