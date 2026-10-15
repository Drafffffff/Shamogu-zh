# Shamogu：萨满山灵守护者

Shamogu v1.6.0 是一款适合短暂游玩的 Roguelike 游戏，强调战术移动，以及谨慎选择召唤图腾灵体和食用食物的时机。视野与噪音构成的潜行机制也十分重要。

宣传短句：图腾 · 潜行 · 净化腐化。

“山中万物正在遭受腐化。野兽逐渐失去理智，变得富有攻击性。这些异变始于一座地城传送门突然出现在山顶之后。作为山之守护者，你决定深入地城，找出并摧毁腐化的源头……愿好运与你同在！”

## 网站

更多信息与游戏技巧请访问[项目网站](https://anaseto.codeberg.page/games/shamogu)。也可以在 [itch.io](https://anaseto.itch.io/shamogu) 网页游玩。另可阅读[《Shamogu：设计随想》](https://anaseto.codeberg.page/games/shamogu/design.html)。

## 从源码安装

请先安装 Go 1.24 或更高版本，确保命令行中可以运行 `go`。Shamogu 使用 [gruid](https://codeberg.org/anaseto/gruid) 网格界面库，支持终端、SDL2 图形界面和浏览器版本；无需单独安装 gruid。

### 终端版（ASCII）

在项目目录运行：

```sh
go build
```

当前目录会生成 `shamogu` 可执行文件。运行 `shamogu -help` 查看命令行选项。终端版使用 tcell 库；支持真彩色的终端可添加 `-t`，以启用与图块版相同的 Selenized 配色。也可以配置终端使用对应颜色；`-x` 选项可作为最后的配色备选。

### SDL2 图形版（图块或 ASCII）

先安装 SDL2（请使用 SDL2，不是 SDL3），再在项目目录运行：

```sh
go build -tags sdl
```

当前目录会生成 `shamogu` 可执行文件。运行 `shamogu -help` 查看命令行选项；例如，`-F` 会以全屏模式启动游戏。

### 浏览器版（图块或 ASCII）

运行以下命令构建 WebAssembly 版本：

```sh
GOOS=js GOARCH=wasm go build -o play-wasm/shamogu.wasm
```

通过 HTTP 服务提供 `play-wasm` 目录即可游玩。该目录还需要 Go 发行版中的 `wasm_exec.js` 文件，详情见 [Go WebAssembly 指南](https://go.dev/wiki/WebAssembly)。

## 文档

命令行选项和回放文件的用法见 `docs/shamogu.zh-CN.6` 中文手册（原文为 `docs/shamogu.6`）。例如，`shamogu -r _` 会自动回放上一局游戏。

## 配置

游戏提供了适合大多数玩家的默认设置。模组选项、按键绑定及常用选项可在游戏内调整；`vars.go` 中还有可在编译时修改的启动选项，部分选项也可通过命令行参数设置。

## Bazzite 与 Steam

SDL2 版的大字界面使用 48×24 文字网格；地图仍保持 80×24 并完整显示。新游戏介绍页采用居中的 80 列布局，以维持原版图案比例。菜单、主灵体和扩展选项采用 48 列大字布局，选项列表在上、说明在下；说明支持 PageUp/PageDown 和鼠标滚轮翻页。启动时不会自动弹出 Steam 虚拟键盘。

## 简体中文翻译说明

汉化覆盖目前审计到的游戏界面、日志、帮助、命令行文案和网页界面，并经过静态字符串审计、Linux x86_64 构建及本地绘制测试。游戏包含随机生成内容和多种模组组合，未逐一核对所有随机分支与状态组合；实际运行中仍可能遇到未译或不通顺的文本，欢迎反馈。

## 素材与字体署名

游戏原版图块素材由 anaseto（Yon）制作，来自官方 itch.io 发行素材，依 CC BY 4.0 许可使用。中文像素字体复用 HarmonistPixelSC；Fusion Pixel 与 Noto Sans SC 字体依 SIL Open Font License（OFL）许可，字体授权文本见 `fonts/` 目录。
