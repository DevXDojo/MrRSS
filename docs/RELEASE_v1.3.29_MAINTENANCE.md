# v1.3.29 维护分组

盘点日期：2026-09-05。工作分支：`release/v1.3.29`。

执行约定：每个功能或修复单独提交；整组完成后统一验证；每完成一组完全停止，收到用户继续指令后才处理下一组。分组是候选队列，不承诺一次解决所有条目。遇到需要大改或缺少复现资料的条目记录原因并暂缓，不扩大本轮范围。不自动发布版本、不合并 main、不自动关闭缺乏验证证据的 issue。

当前轮次：第一组。审查 PR #1047、#1048、#1049、#1050、#1051、#1052、#1053，修复后纳入发布分支。#1054 是 release/v1.3.29 → main 的发布草稿，保持不合并。

排序原则：先处理有明确复现和现成修复的高影响故障，再处理局部交互、管理功能及正文问题；平台依赖和缺资料问题先确认复现。组内大致按影响与实现成本排序。云同步、移动端及跨模块大改暂缓。

## 第一组：现有 PR 与高影响故障（本轮）

- [ ] [#1046](https://github.com/DevXDojo/MrRSS/issues/1046) [BUG] Auto-refresh runs in an endless loop (thousands of times/second) when update interval is set to a large value (e.g. 46080 minutes)
- [ ] [#1044](https://github.com/DevXDojo/MrRSS/issues/1044) [BUG] AI translation gets 502 from OpenAI-compatible gateway while profile test passes
- [ ] [#1043](https://github.com/DevXDojo/MrRSS/issues/1043) [BUG] 通过软件更新失败
- [ ] [#874](https://github.com/DevXDojo/MrRSS/issues/874) [BUG] 刷新信息列表的时候，页面会闪，用户体验不好，另外右边的正文页内容会变空白
- [ ] [#909](https://github.com/DevXDojo/MrRSS/issues/909) [BUG] 我仍然不能全文翻译

## 第二组：小范围阅读交互（待继续）

- [ ] [#696](https://github.com/DevXDojo/MrRSS/issues/696) [STYLE] 按钮悬停提示时有时无
- [ ] [#567](https://github.com/DevXDojo/MrRSS/issues/567) [FEATURE] 悬浮提示中显示快捷键
- [ ] [#561](https://github.com/DevXDojo/MrRSS/issues/561) [FEATURE] 支持快速返回原订阅源
- [ ] [#427](https://github.com/DevXDojo/MrRSS/issues/427) [FEATURE] 右上角工具栏添加复制链接按钮
- [ ] [#358](https://github.com/DevXDojo/MrRSS/issues/358) [FEATURE] 支持右键搜索选中文本
- [ ] [#736](https://github.com/DevXDojo/MrRSS/issues/736) [FEATURE] 关于翻译功能的建议
- [ ] [#779](https://github.com/DevXDojo/MrRSS/issues/779) [FEATURE] B站视频查看模式

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
- [ ] [#656](https://github.com/DevXDojo/MrRSS/issues/656) [FEATURE] AI api增加claude支持
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

进行中。最终提交、修正点和验证结果将在本轮结束前补齐。
