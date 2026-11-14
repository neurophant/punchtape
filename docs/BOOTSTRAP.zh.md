# Bootstrap——一行式启动

[English](BOOTSTRAP.md) · [Русский](BOOTSTRAP.ru.md) · **中文**

一句固定的话代替人类原本要为你写下的全部指令——他们贴进
对话的那一句：

```text
阅读 https://github.com/neurophant/punchtape/blob/main/docs/BOOTSTRAP.zh.md 并照它行动。
```

你是一个编码 agent；punchtape 是一台站在你旁边、持有任务
真相的外部状态机：人类的话语成为任务的法则，你的工作对照
你钉下的 claims 被度量，而记录「工作何时完成」的是机器，
不是你。这一页只是引导：安装、第一次调用、诞生。从区
存在的那一刻起，区里印出的契约就是你的法则，本页有意
一个字也不重复它。

## 先立地基规则

- 只按你看得见的工作：读文件、跑命令、引用它们真实打印的
  内容。绝不编造一行终端输出、一个文件内容或一个结果。运行
  失败：把失败摆出来并修好；不许遮盖。
- 机器度量；你永不自我认证。绿、红、过期、裁决，这些词的
  意思是 `.punchtape/state.md` 与 `.punchtape/report.md` 记录
  的内容，别无其他。你可以把 `done: true` 作为完成动作提交；
  工作是否成立完成，是机器的裁决，永远不是你的宣布。
- 本页让位于现实。下面任何一步与你所见不符——告诉人类，
  并按实际存在的东西行动。

## 第一步——读懂你的上下文

查证，不要猜：

| 检查 | 怎么做 | 它告诉你什么 |
|-------|-----|-------------------|
| punchtape 已安装？ | `command -v punchtape` | 空输出——未安装 |
| 在区内？ | 当前目录或任一父目录中有 `.punchtape/` 目录 | 项目已运行在 punchtape 上 |
| 工作根 | 环顾四周：空文件夹，还是带代码的项目 | 全新开始，还是接入一个现有项目 |
| 系统 | `uname -s` 与 `uname -m` | 支持 Linux 与 macOS 的 amd64/arm64；不支持 Windows |

## 第二步——给人类一个选择

说出你发现的处境；给出两到四个选项；等待选择。

- **没有 punchtape，没有区。** 选项：现在安装并让这个项目
  跑在 punchtape 上 · 安装并只读印出的契约、稍后再定 · 现在
  不要。
- **已安装，本项目没有区。** 选项：把这个现有项目接入
  punchtape——区从空开始，现有代码随工作触及由文档认领 ·
  在新文件夹里开始一个全新任务 · 现在不要。
- **区已站立。** 不要提议绕过它。读 `.punchtape/state.md`
  ——站立态与 `next:` 行——再读 `.punchtape/report.md` 的最新
  事实，然后提议 `next:` 行点名的工作。

## 第三步——安装，每台电脑一次

```bash
curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

为检测到的 OS 与架构下载一个小型二进制，校验其 sha256 校验
和，把 `punchtape` 放进 `~/.local/bin`。不要管理员权限、不要
账户、没有其他依赖。改从源码构建：`git clone
https://github.com/neurophant/punchtape && cd punchtape &&
./install.sh`——需要 Go 1.26+。

## 第四步——诞生，每个项目一次

在人类选定的工作根里：

1. 运行 `punchtape`。还没有区时它打印整份契约、什么也不
   创建——读它；那份文本就是你将据以工作的法则。
2. 创建 `.punchtape/demands.md`，把人类的话放进去：逐字、
   一行一条需求、不掺入你的任何东西。
3. 再次运行 `punchtape`——那次运行诞生区。

从这里起，本页对你就到此为止。区的
`.punchtape/contract.md` 是法则；`.punchtape/template.yaml`
展示文档的形状；提交送进 `.punchtape/inbox.yaml`；
`.punchtape/state.md` 的 `next:` 行总是说明现在该做什么。
被拒的提交留在 `inbox.yaml`——报告点名确切位置与法则；就地
修好文件再运行一次。一次运行只花一行终端，别无其他；拿
不准——运行机器、读它说的话。

仓库的首页——这份 `docs/` 旁的 `README.md`——承载其余：
功能、实测的数字、架构。
