# FAQ——一问一答

[English](FAQ.md) · [Русский](FAQ.ru.md) · **中文**

难的问题都被大声问出、用事实作答。每个答案的纪律：先说出
真相，然后是机制，然后是证据（一个你能打开的链接），然后是
对它正在做或没做的事。「是的，确实如此」在这里是有效回答：
诚实的限制是这台机器设计的一部分，不是需要遮掩的难堪。

## 目录

- [这和……有什么不同？](#这和有什么不同)
- [科学](#科学)
- [理论](#理论)
- [机制](#机制)
- [信任与安全](#信任与安全)
- [产品](#产品)
- [一句话质疑](#一句话质疑)
- [关于呈现](#关于呈现)

## 这和……有什么不同？

**……和 CI 或 pytest？**
CI 事后评判一次 push，裁决由人来读；它从不与写代码的 agent
说话。punchtape 工作在任务循环之内：它持有任务的覆盖（人的
言语每一行都被申报、树的每一行都被 claim 或被公开划掉），
在每次改动上亲自运行检查，把失败回喂给 agent，并用校验和
封存完成。pytest 套件只是检查可以烧的一种燃料：你的套件
留下，并且在机器之下变得更强。
证据：[USAGE.zh.md](USAGE.zh.md)、[演示](../demos/)。一份
在任何 push 之前就已记录在案的裁决。

**……和 harness（Claude Code、Cursor……）？**
harness 运行你助手的会话、活在 agent 的环境里——它就是工作
的当事一方。punchtape 是持有任务真相的外部一方：言语、证据、
裁决。harness 原封不动；仲裁者加在旁边。
证据：机器从不读你的聊天、从不问 agent 任何事——
[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，「需求流」。

**……和 spec-kit 这类规格驱动工具？**
那些工具帮 agent 在自己的上下文里写 spec、持有 spec；spec
和它的检查仍是 agent 自己的话。这里，人的言语被一台外部
机器逐字取走，检查被钉到代码行上、由机器执行，完成也由
机器封存，agent 无法给自己的作业签字。
证据：需求追加记录与链——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，
「需求流」。

**……和 guardrails 与沙箱？**
guardrails 约束 agent 可以做什么；沙箱约束它在哪里做。两者
都不测量做完的工作是否达到主人的言语。punchtape 做的恰好是
这一件事，别的都不做——并且与你已在用的任何沙箱叠加。
证据：裁决闸门——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，
「裁决」。

**……和「写测试就行了」？**
测试是燃料——机器运行它们、钉住它们的字节、在它们 claim
的行移动时重跑，并且在任务或树还有任何一行未了结时扣住
完成。写测试是你的 agent 在机器之下做的事；让测试保持诚实
是机器替你做的事。
证据：[FEATURES.zh.md](FEATURES.zh.md)，claim 法则与过期
法则。

**……和 make/just？**
make 按你的清单执行，一次、在你开口时；它不持有言语、不持有
覆盖、不持有裁决、不持有历史。punchtape 是一台状态机：它知道
哪些检查过期了、你言语的哪些行未了结，并且在还有任何开放
之物时不记录 done。
证据：运行的过期算术——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，
「运行」。

**……和 linter？**
linter 按给所有人定的固定规则查风格。punchtape 对照你的言语
查这个任务的行为——检查为你的任务而写、钉在你的行上、改动
即重跑。

**……和 agent 自己的自审？**
自审就是 agent 给自己的作业打分——委托代理问题浓缩成一个
动作。这里 agent 可以写检查，但运行它们的是机器、钉住字节
的是机器、记录裁决的是机器；自己宣布的「done」是机器亲自
复核的一行裸文。
证据：完成动作——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，
「裁决」。

**……和 IDE 的 watch 模式？**
watch 模式在文件变化时重跑一条固定命令。它没有你的言语、
覆盖、封存完成的概念；而且它跑在你的编辑器里，不是为你的
agent 跑的。

**……和 AI 代码评审（CodeRabbit、Greptile……）？**
评审者读 diff、留评论；读的人仍然是你，环仍然开着。
punchtape 不评论、直接运行：检查在树里执行，失败作为工作
退回给 agent，完成被扣到每一道闸都立住为止。评论可以无视；
被扣住的裁决不能。
证据：运行与裁决法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

## 科学

**论文是预印本——同行评审在哪里？**
没有；它是预印本，自发表于三个带 DOI 的预印本服务器。在那里
发文是付费托管：买的是一个位置和一个永久 DOI，从来不是
评审；三个服务器是一份文本的镜像，不是三篇论文。它背后
站着的是：方法论与 campaigns 的原始数据在本仓库公开，这里
引用的每个数字都能追到你打得开的文件。
证据：[RESEARCH.md](RESEARCH.md)、[METRICS.zh.md](METRICS.zh.md)、
[campaigns/METHODOLOGY.md](../campaigns/METHODOLOGY.md)。

**作者评估了自己的工具——利益冲突？**
是的，这个冲突真实存在，并且已声明。设计回应了它：验收由
盲评裁判对照人的原始言语测量，从不由机器测量；每任务的
原始数据公开，可供独立重算。
证据：已声明的[利益冲突](RESEARCH.md#conflicts-of-interest)；
裁判角色——[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement)；
按密封清单导出的盲评样本。

**「盲评裁判」是谁——你雇的？**
没有人被雇。裁判是一个由模型在冻结提示词下执行的角色：每个
cell 一个微裁判、cell 的完整工作树、对照任务自身的输入
打分，从不看组别标签，从不看机器自己的裁决。该角色在两组共
评了 1720 个点，0 个缺失。
证据：[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement)；
计数——[METRICS.zh.md](METRICS.zh.md)。

**n=25 个任务——置信区间在哪里？**
这是实地评估，不是临床试验。方法论确实固定了机器：各波
之前的两阶段预注册、比例用 Clopper–Pearson 区间、差值用
bootstrap、多重比较用 Holm 校正。这里展示的是测量层：每个
任务按系列留档（RESEARCH.md 的附录）、每任务的原始指标在
文件里。请把标题数字当作数据开放的实测
田野结果。
证据：[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement)，
Statistics；[RESEARCH.md](RESEARCH.md)，
各系列附录。

**25 个全过——循环论证？验收由机器定，当然全过。**
验收不是机器说了算。机器记录事实；盲评裁判把结果与人的
原始言语对照。机器读不了含义；那是裁判的工作（也仍然是
你的）。
证据：[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement)；
裁决闸只测形式完整——
[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，「裁决」。

**任务是作者挑的——cherry-picking？**
问题台账法则禁止：每个系列里进入的每个任务都连同输入与
结果站在台账里；没有任务被丢弃。
证据：[METHODOLOGY 0.2f](../campaigns/METHODOLOGY.md#02f-the-problem-ledger-law)，
问题台账法则；台账与原始数据放在
一起。

**没有和 CI/spec-kit/竞争对手的对比？**
属实——基线是同一批工作不用 punchtape，而不是与其他工具的
正面对比。这是主张的诚实边界：数字说的是相对「无仲裁者」
改变了什么，对对手一个字都没说。
证据：[RESEARCH.md](RESEARCH.md)，分组设计。

**霍桑效应——被看着的人表现更好？**
隔离法则按 campaign 分组：会话、权限、材料在程序上完全相同，
裁判看到的是匿名导出。残余的观察者效应无法排除，也没有被
硬说不存在。
证据：[METHODOLOGY 0.2](../campaigns/METHODOLOGY.md#02-the-campaign-isolation-law)（隔离）、[I.8](../campaigns/METHODOLOGY.md#i8-validity)（效度）。

**成本和 token 是作者自己记的？**
按每任务指标文件法则：每个任务带自己的指标文件——token、
时间、钱，公开在原始数据里，谁都能重算。
证据：[METHODOLOGY 0.2e](../campaigns/METHODOLOGY.md#02e-the-per-task-metrics-file-law)；
campaigns/ 下的原始文件。

**5–19× 的经济学靠编造的时薪？**
方法固定了工时怎么记账（机器经济账，
[METHODOLOGY I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy)），
给出的是跨实地任务的一个区间，不是一个点。对费率的敏感
正是区间宽的原因：这是诚实，不是注水。
证据：[METHODOLOGY I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy)；
[METRICS.zh.md](METRICS.zh.md)。

**论文是 LLM 写的？**
事实是机器可查的：胶片、日志、校验和、开放数据；文本的
风格哪边也证明不了，也不被当作任何事的证据。

**裁判是 LLM——LLM 当裁判的脆弱有据可查。**
属实，而且公开说：裁判角色由模型执行。设计以这份脆弱为
前提。裁判逐行执行一张固定的表，字节级严格，无权编辑或
变通。疑点计入产品不利。任何计数之前必须先通过校准，每张
表在一波之前验证两次——先黄金产品，再单功能坏掉的产品。
证据：[RESEARCH.md](RESEARCH.md)，"Judges"；
[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement)。

## 理论

**古德哈特定律：agent 会钻检查的空子，而不是钻含义。**
会——检查是字节，agent 可以写一个空壳（`echo ok`）。设计
收窄这个缝而不是否认它：检查钉在 claim 的行上（确切的期望
输出随行记录）、站在 spec 页上——每个行为一页、每条检查
与你的言语并排展示——豁免带引用的理由、响亮，而含义的
最后一眼留给人。
证据：spec 页与 claim 法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
[FEATURES.zh.md](FEATURES.zh.md)。

**预言机问题只是搬进了 YAML？**
言语的第一句话仍出自人：那个预言机是你，没有工具能移走
它。从你身上移走的是预言机的执行：便宜、可重复、字节精确、
写下之后不可变。
证据：需求流及其链——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**这是 verification 还是 validation？**
Verification。机器查形式完整：言语全部了结、每条 row 以钉住
的字节保持绿色、树等于 claims 加响亮的豁免。这是否符合你的
本意——validation——是人的最后一眼，README 里说得直白。
证据：[README](../README.zh.md) 的诚实限制。

**「proof」和「signature」对校验和来说太隆重了。**
它们是 canonical 字节上的 sha256 和：要么重算得出来，
要么失败。没有哈希之外的密码学、没有身份声明，文档用这两个
词指的恰好就是这件事。
证据：链与摘要法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**检查是 agent 自己写的——委托代理问题没解决？**
写是 agent 的；运行、钉字节、出裁决是机器的，而每条检查
挨着它的 claim 那一眼是你的。自我认证被拆到两侧；没有宣称
解决——它被做成了可见且难以伪造。
证据：执行与裁决法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**足够详细的 spec 就是代码——你重新发明了编程。**
这话有真相：完全的精确就是编程。机器从不要求言语完整。你的
行保持散文；每个行为 agent 补一条小检查（一条命令和它的
精确字节），由机器运行，在 spec 页上与言语并排。散文钉不住
的地方，诚实的形式是带理由的响亮豁免，而不是假装精确。
证据：文档法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
一个真实 corpus 的形状——[demos/gouija](../demos/gouija/)。

**agent 会靠全部重写来满足 spec——没写下来的行为就丢了。**
会：从没写下的言语保护不了，没有工具能改变这一点。机器对
这道缝做的事：树的每一行必须被检查 claim 或带理由地响亮
划掉，claim 过的行在任何改动后重新挣回钉住的字节，之后坏
掉的完成被标记为 open、连同弄坏它的 run。洞被写下，从不
沉默；含义的最后一眼仍是你的。
证据：覆盖与漂移法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**agent 会自己改测试——reward hacking 有据可查。**
在独立评测里有据可查，在这里也可能发生。结构让它变得响亮：
改拼一条检查等于一条新 row，旧 row 仍然站立、仍然钉着
自己的字节。退役任何检查都需要一份点名人的言语的 mandate。
每条 row 的字节和 run 编号都入账，前后都有，所以削弱一道
守卫会留下书面痕迹。
证据：合并与 mandate 法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

## 机制

**agent 写的 YAML 很脆。**
它是场上最常见的拒绝：含「: 」或逗号的值加引号、制表符被拒、
多行钉保持一行。每个拒绝点名自己的文件、行和法则；被拒的
submission 留在 inbox 里，数据不丢。
证据：文档法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
实地演练的数字——[RESEARCH.md](RESEARCH.md)。

**agent 会写垃圾检查——全绿。**
空壳检查它写得出来；机器读不了含义。机器做的是：把检查钉
在它 claim 的行上、放到 spec 页上人的言语旁边——空壳一眼
可见，而它 claim 的行永远是它的责任。
证据：spec 页——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**这就是 make 加 YAML——重新发明的 pytest。**
pytest 不持有的东西：你的言语作为任务的法则、树被 claims
或响亮豁免覆盖、每条 claim 行的过期算术、带校验和的封存
裁决、比任何会话都长寿的胶片。pytest 想用的时候，作为燃料
在这台机器里跑。
证据：运行与裁决法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
一张活胶片——[demos/enimda](../demos/enimda/)。

**任意命令——rm -rf？**
命令只看得见 `PATH`、`HOME` 和区的 pin 值；它从树根、经
shell 运行，随它的进程组一同死去，整套东西离线工作。跑什么
由主人的言语决定，从来如此。
证据：你的树法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)、
[BOOTSTRAP.zh.md](BOOTSTRAP.zh.md)。

**日志坏了——区死了吗？**
撕裂的尾部自我隔离到旁车文件，完好的记录照常加载；别处的
损坏被拒绝、点名文件与记录，从不被悄悄吸收。区放进 git：
任何绿色状态一条 checkout 就回来，在手册里验证过。
证据：撕裂尾部法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
带真实运行的处方——[USAGE.zh.md](USAGE.zh.md)。

**协议会变——旧胶片烂掉？**
由另一版本机器诞生的区被拒绝——版本从不静默混用。胶片是
文件：归档的区作为记录仍然可读，即使新区从零开始。
证据：诞生与版本法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**NFS 和锁？**
区锁是本地文件系统上由内核释放的 flock；在 NFS 上，安全的
答案是：让区留在本地。
证据：锁法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**没有 Windows？**
今天没有——Linux 与 macOS，amd64 和 arm64。移植是开放的
工作，不是设计上的拒绝。
证据：[README](../README.zh.md) 的系统要求。

**一个导入（yaml）和 Go 1.26——依赖呢？**
机器里一个外部导入，带着自豪说；Go 1.26+ 只为从源码构建——
发布版二进制只需要一个 unix shell。
证据：[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，包地图。

**通过 demands 或报告注入提示词？**
机器从不执行散文：言语是逐字取走的行、文档是经严格 decode
校验的数据、row 只跑它们的 pin 所说的。报告里被注入的一句话
是你读到的文本，机器不据它行动。
证据：文档与执行法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**agent 能直接改 .punchtape/ 文件。**
能碰；机器随后拒绝：链在字节上重算，记录要么重新合成、要么
区已损坏——响亮、点名、exit 1。篡改按设计可检测，而不是靠
权限阻止。
证据：链法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
活的拒绝行——[USAGE.zh.md](USAGE.zh.md)。

**格式上的 vendor lock-in？**
区是纯文件（markdown、YAML、以 sha256 命名的对象），离开
机器也能读、`cp` 就能复制、git 就能 diff。胶片的可读性不
依赖这个二进制。
证据：区的文件清单——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**团队流程——区在仓库里还是外面、冲突呢？**
两种形态都带着处方和真实的拒绝行走过一遍：什么入库（胶片）、
什么保持生成（`cache/`、`derived/`、`run.lock`）、合并怎么
解（一侧完整保留、日志永不合并）、坏区怎么治（把字节放
回去）。
证据：团队手册——[USAGE.zh.md](USAGE.zh.md)。

**curl | bash——把脚本管道进 shell？**
有道理：把没看过的代码管道进 shell 是真实的风险。脚本就在
本仓库——跑之前先读。它校验发布校验和、把一个二进制装进
`~/.local/bin`，不要管理员权限。不想用管道：手动下载二进制
和校验和，或从源码构建。
证据：[get-punchtape.sh](../get-punchtape.sh)；
[BOOTSTRAP.zh.md](BOOTSTRAP.zh.md)。

**agent 是机器的操作员——狐狸看鸡窝。**
是：操作员是 agent，而一个控制你整个屏幕的 agent 能伪造你
看到的一切。所以真相活在机器持有的文件里：日志每次运行
重算它的链、视图从记录重新生成、replay 逐字节重建胶片、区
可以和代码一起放进 git。任何设计都给不了你一个不会说谎的
agent——这正是裁决页为你的眼睛而写的原因。
证据：链法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
replay——[演示](../demos/)。

**检查者本身是 AI 吗——它会幻觉出裁决吗？**
机器里面没有模型。row 是命令；机器运行它们、比较字节；同样
的输入在每台机器上给出同样的答案。一条会抽风的命令是任务
自己的问题——死线杀掉挂起、事实记成红色，而机器自己的
答案从不变化。
证据：[FEATURES.zh.md](FEATURES.zh.md)，「确定性的运行」；
[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，「过期与执行」。

**100% 函数覆盖——虚荣指标？**
它是地板，对它声明的东西仅此而已。这个数的意思是：机器的
每个函数在构建的测试套件下至少执行一次——测试加上在内置示例
任务上跑的插桩二进制。语义重量由另外两道闸承担：示例必须
穿过真实的机器闭合它的裁决，并且每条被强制的规则都必须
由契约打印出来——同一个集合，双向核对。
证据：构建时的闸门——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**长任务让上下文烂掉——状态活不过它们。**
上下文是最先烂的，属实。状态活在所有上下文之外：standing、
corpus、教训和下一条指令都是文件，终端里的那行就是一行。
会话或模型可以在任务中途更换——下一次运行读的是区，不是
聊天。演示展示了工作跨运行、站在已封存裁决上继续。
证据：日志与视图——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)；
[demos/enimda](../demos/enimda/)。

## 信任与安全

**怀有恶意的 agent 能破坏区。**
真——有文件权限的 agent 能烧掉胶片。它做不到的是伪造一份
standing 裁决：链和记录重算，破坏以响亮损伤、点名文件的
形式现身。把区放进 git，连破坏都可回滚。
证据：链法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**胶片事后能伪造吗？**
每份裁决的 signature 是 canonical 记录字节上的 sha256；改动
任何东西，下一次运行总和就对不上。伪造者必须一致地重算每
个哈希——这道检查机器替读者做了。
证据：折叠与签名法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**被调包的二进制可以对裁决说谎。**
会——信任这个二进制，或者从源码构建；发布带校验和，安装器
会验证。无论如何，胶片都是独立可查的文件。
证据：发布校验和；[BOOTSTRAP.zh.md](BOOTSTRAP.zh.md)。

**凭什么信任这台机器？**
因为它要的是可给出的最小信任：读胶片（纯文件），再跑一遍
replay；它声称的一切都是你查得到的字节。演示在你机器上
从零做的正是这件事。
证据：[演示](../demos/)——每个都带 `replay.sh`。

## 产品

**才 0.5——现在依赖太早？**
有道理。协议年轻、版本拒绝混用，这是封存过去的诚实代价。
支撑早期采用者的是：开放的代码、开放的数据、比二进制长寿的
纯文件胶片。

**这怎么赚钱？**
MIT、无账户、无遥测——机器里没有收入机制。以后是什么是
开放问题，等有东西可答时再答。

**路线图是什么？**
Windows、围绕区的团队便利、集成——活还是死由真实用户的
issue 决定——胶片决定什么重要。

**巴士因子是一？**
今天属实。缓解是结构性的：开放的方法论、开放的原始数据、
纯文件胶片、MIT——工作活得比作者久。

**为什么是 CLI 而不是服务？**
因为仲裁者必须站在每个会话和每个账户之外：一个在你项目
旁边的二进制、离线、没有要信任的服务器、没有会泄漏的数据。
服务会是又一方——这台机器存在的目的恰恰是不做那一方。
证据：设计法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**和我的编辑器/CI/agent 集成？**
机器在法则上不可知——任何能跑 shell 命令的 agent 都已经能
驱动它；CI 可以在 push 后重放一个区。演示展示四种栈；你的
harness 是第五种。
证据：不可知法则——本仓库的 AGENTS.md；[演示](../demos/)。

**harness 们会把它建进去——你为什么不会被吸收？**
功能会被吸收，软件就是这样走的。吸收不进 harness 的，是
这台机器为之存在的那件性质：站在会话之外。内置的检查器又
是 agent 给自己的作业打分——本项目为之存在的自我认证正是
要打破的东西。一个外部的、不可知的、离线的二进制与每个
agent 协作，今天的和明天的；如果 harness 们有朝一日给出
基于开放胶片的封存外部裁决，用户就赢了。
证据：设计法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)，
「机器是什么」。

**谁在生产环境用它？**
不宣称有生产用户。今天敞开的是：三轮带每任务原始数据的
实地测试，和四条谁都能从零重放的演示胶片。这句话会在它
不再是真话的那天改变。
证据：[RESEARCH.md](RESEARCH.md)；[演示](../demos/)。

**演示是你们 linter 之外的引入代码。**
对 Makefile 的诚实读法：引入的演示项目切片被排除在 lint
闸之外，就一行可见的命令、带理由。机器自己的代码——cmd/、
internal/、tools——在完整的测试套件下运行。演示带着真实项目的
钉住切片作为胶片的材料，每个都标着来源 commit，每张胶片由
它的 replay 再生。
证据：Makefile；每个演示的 README 及其钉住的来源。

## 一句话质疑

**「模型变好了——问题没了。」**
论文的对照组用的是同级别的模型，25 个任务里产生了 32 次
人工验收；更好的代码不会自己检查自己。等它会了，带着我们
的祝福埋掉这个工具。
证据：[METRICS.zh.md](METRICS.zh.md)。

**「自己检查一下就行了。」**
那就是基线，实测的：每 25 个任务 4.4–7.3 小时。机器组：
0.8–1.3 小时，25 个任务 25 个首次提交即完成。
证据：[METRICS.zh.md](METRICS.zh.md)。

**「token 开销会把省的吃掉。」**
场上数过：含机器时任务总成本低 5–19×，token 计入每任务的
指标文件。
证据：[METHODOLOGY 0.2e](../campaigns/METHODOLOGY.md#02e-the-per-task-metrics-file-law)、
[I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy)；
[METRICS.zh.md](METRICS.zh.md)。

**「这是 1999 年的 TDD。」**
TDD 从未把人的言语当作不可变的链持有、从未 claim 过树、
从未封存过裁决；这是由外部一方闭合的环：1999 年留给人的
那一半。
证据：链与裁决法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**「又一层官僚。」**
你现在的官僚是你自己：一遍遍重读 diff。这里加的这层把那种
阅读交给机器去做；你写下的言语就是全部界面。
证据：[METRICS.zh.md](METRICS.zh.md)，验收时间各行。

**「离线——云都不行？」**
离线就是特性：仲裁者把任务的真相持在你的代码旁边，账户、
遥测、网络调用都碰不到它。
证据：你的树法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**「『机器』这个腔调是邪教。」**
这个口吻是刻意选择并保持一致的，为了让术语只有一个意思；
如果读起来像仪式，词汇表把每个词都翻译成它朴素的文件与
字节含义。
证据：[GLOSSARY.zh.md](GLOSSARY.zh.md)。

**「今天 MIT，明天专利？」**
今天 MIT。没有申请、没有计划；真有那天，已授予的许可证也
收不回去。

## 关于呈现

**为什么 README 是这个样子？**
它按 问题 → 方案 → 证据 → 演示 → 上手 构建，每条限制都摆在
首页；深度住在 USAGE。哪一节不再卖东西、开始布道，那就是
bug，开 issue 说。

**为什么叫这个名字、为什么是带子？**
穿孔纸带：一个孔，一旦打上，就无法撤销；日志给每条记录
在案的事实同样的性质。
证据：链法则——[ARCHITECTURE.zh.md](ARCHITECTURE.zh.md)。

**为什么词汇表读起来像法典？**
因为每个术语指一个机制、一条法则；松散的同义词正是测量的
东西腐烂的方式。一个词，一个含义，一个文件。
证据：[GLOSSARY.zh.md](GLOSSARY.zh.md)。
