# 用法——每个场景，一步一步

[English](USAGE.md) · [Русский](USAGE.ru.md) · **中文**

这一页是给持有任务的人看的。你的助手有自己的入口：
[BOOTSTRAP.zh.md](BOOTSTRAP.zh.md) 负责第一次接触；一旦有了区，
`.punchtape/` 里还有打印出来的契约。那份契约是法则；这一页
从不取代它。下面的一切归结为一个技能：你用文字写任务；其余
由机器和 agent 完成。

## 安装——每台电脑一次

把这段贴给你的 AI 助手——或者自己在终端里运行：

```text
安装 punchtape 并让所有工作都经过它：

  curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

它做什么：下载一个小二进制，校验其校验和，把一条命令
`punchtape` 放进 `~/.local/bin`。不要管理员权限，不要账户。
每台电脑做一次，从不需要每个项目做一次。（改为从源码构建：
`git clone https://github.com/neurophant/punchtape && cd punchtape && ./install.sh`——需要 Go 1.26+。）

再加一行，只需一次：如果你的助手有常驻指令（你给它写长期
规则的地方），加上「让所有工作都经过 punchtape」这一行——
以后你再也不用在提示里写它。

什么时候该把「让所有工作都经过 punchtape」加进提示——一条
简单的规则：

- 文件夹已经在跑 punchtape（里面有 `.punchtape` 文件夹）？
  → **不需要。**那个文件夹里的 `AGENTS.md` 已经告诉每一个
  AI 工具这件事了。
- 你曾经把这一行写进助手的常驻指令（大多数助手都有）？
  → **不需要——任何地方、任何时候都不需要。**
- 新的空文件夹，而且助手没有这类常驻指令？→ **加上这一行。**
- 拿不准？→ **加上这一行。**它从不坏事。

## 开始任何新项目——第一个和第十个一样

打开一个新的空文件夹。用平常的话把任务写给助手：

```text
做一个笔记工具：notes add <文本> 保存一条笔记并打印它的编号；
notes list 每行一条、从旧到新打印全部笔记；notes show <编号>
打印一条笔记；状态存在 notes.json 里，重启之后依然存在。
```

接下来发生什么，一步一步（下面的终端行都是这一版本的真实
输出）：

1. 你的助手在文件夹里运行 `punchtape`。文件夹是空的，于是
   机器打印自己的契约——整整一页机器的全部法则：做什么、写
   哪些文件、按什么形状。助手读完就知道了整个流程。你什么
   都不用解释。
2. 助手把你的任务文本——你的原话，一字不改、一行一句——
   放进 `.punchtape/demands.md`。下一次运行诞生区：机器的
   完整手册（契约、模板、参考页）落在你的言语旁边。从那一刻
   起，你言语的过去就被锁住了：被碰过的过往行就是区的损伤，
   机器会停下。新的言语随时可以继续进同一个文件。

   ```
   punchtape: run 0 — corpus empty — submit the first document as .punchtape/inbox.yaml (copy .punchtape/template.yaml)
   ```

3. 助手把 `.punchtape/template.yaml` 复制为
   `.punchtape/inbox.yaml` 并填写：提交的故事、为你言语每个
   新行申报的结果、要求和检查——每个行为 claim 确切的代码
   行，每条检查是一条命令加它钉住的期望输出。机器亲自运行
   rows，并用一行说出你和完成之间还隔着什么：

   ```
   punchtape: run 1 — behaviors without a green pinned row: notes/list — submit to .punchtape/inbox.yaml a row whose expect pins its exact output
   ```

   失败的检查落进 `.punchtape/report.md`——期望什么、实际发生
   什么，并排放着，以字节计。助手读这个文件，修代码或修
   rows。
4. 修复落下；机器只重跑过期的东西，并点名剩下的洞：

   ```
   punchtape: run 2 — 10 uncovered lines — submit their claims or loud exclusions to .punchtape/inbox.yaml
   ```

   工作盖不住的文件得到响亮的豁免——公开、带理由、绝不
   静默。
5. 全部检查绿色 → 终端说出来，助手往 inbox 里写整整一行：
   `done: true`。这就是签收——不用重打，不用排版。

   ```
   punchtape: run 2 — all green — write done: true to .punchtape/inbox.yaml and run to record the verdict
   ```

6. 机器重新检查一切：每条 row 以钉住的 proof 保持绿色、你的
   言语整体了结、每条要求被覆盖、树等于 claims 加响亮的豁免、
   没有开放的问题——然后记录裁决：

   ```
   punchtape: run 2 — verdict #1 standing — show .punchtape/state.md to the human; new work goes to .punchtape/inbox.yaml as a document
   ```

   从此 `.punchtape/state.md` 显示：`verdict: #1 … —
   standing`，连同它的校验和。谁问它管不管用，就把这个文件
   给他看。

之后的下一个新项目还是同样的事：空文件夹、文字任务、完成。
每个文件夹各自独立；一个项目里的东西不会漏进另一个。

## 继续一个已经跑在 punchtape 上的项目

打开那个项目的文件夹——或它里面的任何文件夹（机器向上
查找、找到自己的区）。用文字写下你要什么：

```text
加一个 open 命令——只显示未关闭的笔记。
```

谁来接手都无所谓：你惯用的助手、新会话、另一个 AI 模型、
同事。项目根目录的 `AGENTS.md` 里有一段 punchtape 留下的
短记——每个 AI 工具到场都会读它——短记指向
`.punchtape/state.md` 那一页：工作站在哪里、下一步是什么。
任何人都能在几秒内接上。

你的改动会发生什么：你的言语并入 `.punchtape/demands.md`；
claim 了被改文件的检查重新运行；仍然成立的检查不重跑、不
花钱。改动可行之后，机器记录完成 #2——然后 #3、#4……每份
裁决都是自带校验和的记录在案的事实；若之后的改动弄坏了
已封存的结果，机器会点名抓住它的 run 和修复它的 run。

## 真实情境的真实提示

**一个 bug。**
`notes show 3 在没有 3 号笔记时什么也不打印——修好，并证明它不可能再发生。`
你的助手为想要的行为加一条检查（它失败——bug 存在），修
代码，检查通过，而这条检查从此永远留在项目里：它 claim 的
行一动，它就重跑。[demos/](../demos/) 里的 rjq 演示就是这
个故事，发生在一个现成的 Rust 库上。

**一个从没用过 punchtape 的现有项目。**
`把这个仓库纳入 punchtape：读代码，为已经存在的行为写检查，让它们全部通过。`
机器测量你的代码今天真正做什么，并替其余部分起草豁免
清单。从那以后，每次改动都对着它测量。gouija 演示是一个
现成 Go 仓库上的这个故事；enimda 演示展示随后到来的
那个功能。

**一个新的 AI 模型接手了。**
`通过 punchtape 继续这个项目——先读 .punchtape/state.md，然后补完缺失的部分。`
新模型不需要旧聊天：项目文件夹里什么都有。

**把项目交给同事。**
复制文件夹。完事。言语、检查、结果、历史、完成——全都在
里面，逐字节相同。

## 团队手册

### 区在仓库里

把胶片整个提交：`demands.md`、`journal/`（epoch 文件与
对象）、`corpus.yaml`、`state.md`、`report.md`、`spec/`、
`holes.md`、`trace.md`、`knowledge.md`、`contract.md`、
`template.yaml`、`reference.md`，以及机器在根 `AGENTS.md` 里
的标记块。这是任务的记录在案的过去——放进版本控制，它熬得
过一切，一个克隆立刻携带全部记录。

生成性的服务状态别进 git：`.punchtape/` 里的 `cache/`、
`derived/`、`run.lock`——机器每次运行都会重建它们。
（`.punchtape/tmp/` 不需要自己的规则：机器在每次运行结束时
清空它。）新同事不需要入职引导：克隆下来，他的 agent 读
`AGENTS.md`、打开 `.punchtape/state.md`、顺着 `next:` 行指的
地方继续。

### 区在仓库外

没有什么强迫区进 git。`.punchtape/` 进了你的 `.gitignore`：
仓库携带代码，区携带真相，两边各自完整。每个开发者（和你的
CI）看到的是纯树。任务交接时转移的是文件夹，不是分支。区们
分叉时，真相住在机器自己的记录里：日志——改它一下机器就会
点名。代价同样直白：没有区的克隆从空开始，所以这个形态更
适合单一主人或明确的交接，而不是一大群人。

### 冲突与处方

机器遇到每个冲突都以响亮的拒绝加点名文件回应——它从不猜、
从不静默吸收损伤。下面每一行都是这一版本的真实输出。处方
是你贴给助手的行；动手的是 agent。

**一个运行站着的时候来了并行调用。**运行按区自己的锁串行
（`run.lock`，由内核释放、死亡安全）：第二个调用排队等，
轮到它就继续——实测过：一个外部持锁者把一个 run 恰好延迟了
它持锁的那么久，随后该 run 正常完成。你什么都不用做；锁
不归你管。

**日志在合并中冲突了（两个克隆、两边都提交了）。**日志是
一条线性记录：它永不合并。非尾部的损伤被拒绝、点名文件与
记录：

```
punchtape: journal: record 0 of /tmp/epoch-conflict/.punchtape/journal/epoch-00002.yaml does not parse: [1:17] could not find flow map content
```

撕裂的最后一条记录不一样：机器把它隔离到完好记录旁边的
`<epoch>.torn`，然后继续。处方：

```text
解决 .punchtape 的合并冲突：完整保留恰好一侧——demands.md、journal/、corpus.yaml 整体保留，绝不混装。然后运行 punchtape。如果它报告损伤，从区还是绿色的最后一个提交恢复 .punchtape（git checkout <sha> -- .punchtape），并把另一侧的言语和文档作为新工作重新提交。
```

**言语的过去被碰了。**链在每次运行时按字节重算，被改动的
过往行让一切停下：

```
punchtape: zone damaged: the demand stream's past changed
```

放回精确字节就能修复——实测过：同一个区、同一次运行，原始
字节回来后重新 standing。区在 git 里的话，恢复就是一条
checkout：

```text
.punchtape/demands.md 被手工编辑过，机器报告 zone damaged。从 git 恢复该文件的精确字节（git checkout -- .punchtape/demands.md），然后再次运行 punchtape。
```

**一个由另一版本机器诞生的区。**版本不混用，整份处方机器
自己说：

```
punchtape: zone born by machine 9.9.9 — this machine is 0.5.1; archive it (cp -r .punchtape .punchtape-archive) and start fresh
```

**树里坐着一个垃圾或半死的区。**没有日志、也还没有言语的
`.punchtape` 文件夹什么都不是——机器这么说完就等：

```
punchtape: no instance yet — put the human's words verbatim into .punchtape/demands.md and run again
```

如果日志没了而言语还在，下一次运行会让区以空白状态重新诞生——记录在案
的过去就是日志；没有它，就没有可恢复的过去。处方：

```text
这个 .punchtape 是实验留下的垃圾。把它归档（mv .punchtape ../punchtape-junk），如果这个项目应该跑在 punchtape 上，从我的言语重新开始。
```

区在仓库里时，「重新开始」有一张安全网：胶片在 git 历史里，
它的任何绿色状态都能用一条 checkout 从那个提交取回
`.punchtape`。

**一个树上两个区。**区是工作目录上或之上的最近一个
`.punchtape`——嵌套文件夹会静默遮住项目真正的区。实测过：
从深层的子目录调用，机器指向的是 `../../../.punchtape/state.md`
——上面那个区。处方：

```text
找到嵌套的区（find . -name .punchtape -not -path './.punchtape'），告诉我哪一个是项目的区，归档另一个，并在工作根目录运行 punchtape。
```

**多个 agent，一个区。**锁串行化的是运行，不是工作。谁提交
哪一部分是你作为任务主人的决定：把工作拆成互不重叠的部分，
一次一份提交。punchtape 是事实的仲裁者，不是 agent 的
协调者。
