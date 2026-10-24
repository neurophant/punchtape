contract.md 的镜像；以英文原文为准。

punchtape — 契约

你刚刚运行了 punchtape，它打印了本页：此目录中还没有实例。
什么都不存在，也没有创建任何东西。如何开始：创建文件
.punchtape/wish.md，把人类的愿望逐字放入其中，然后再次运行
punchtape——那次运行诞生实例、永远固定目标并写出本区的现行
文件。

调用，永远。只有一次调用，无参数：`punchtape`。终端只收到
一行——即下一步该做什么，并点名要动的文件。此后的每个答案
都在本区文件里：state.md 载有现状与 next: 指令，report.md 载
有最近一次运行的事实。

循环，诞生之后。提交文档 → 其行执行 → 修代码或修行直到
全绿 → 把 corpus.yaml 逐字节复制为 inbox.yaml 并运行：这次提交
即完成动作，机器记录裁决。此后 state.md 显示裁决与其后的一切
漂移。在每一个节点，终端行与 state.md 的 next: 行都明确说明
现在该做什么。

law:wish      人类的愿望逐字写入 .punchtape/wish.md。下一次运行
              诞生实例并永远固定目标；该槽位在诞生后即告失效。
              没有愿望就没有诞生；没有任何别的东西会创建实例。
              诞生时机器按文件顺序为愿望的每一非空行编号
              W1..Wm——号码永远固定：愿望的字节不再改变。

law:document  全部工作只有一个文档（形状见下）。复制
              template.yaml，编辑，作为 .punchtape/inbox.yaml
              提交。提交按 id 合并：所列行为被新增，或其头部
              替换现行者而其行按 run 合并——同一 run 整行替换，
              新 run 追加，未列出的行保持原样。要求按 id 同样
              合并；提交携带的 not-a-requirement 列表或
              artifacts 列表整体替换现行列表。未列出者原样
              保留；提交行为的子集是正常提交。提交次数不影响
              价格：所列字节只走一次，未改动的行不再重发。
              删除只能是显式的 remove 列出现存 id——行为或
              要求；退订一条要求时，同一提交必须撤下对它的
              全部 covers。notes 为自由文本，永不被解释。

              登记簿为愿望记账：要求以 W 区间点名所占的简报
              行；行为在 covers 中点名自己证明的要求；不提出
              要求的行以 not-a-requirement 标注——逐字引用
              该行自己的文字，旁附理由；artifacts 声明工作的
              产出。每一行都携带显式
              expect——rc、out 或 files——且每个行为至少带
              一行：行为证明的一切就是行。

law:paths     一切路径——expect 文件、scope 通配、artifacts——
              都相对工作根。绝对路径、「..」与越界写法一律
              拒绝。

law:done      提交语料自身的字节（corpus.yaml 原样复制为
              inbox.yaml）即记录裁决的完成动作。裁决需要：
              每一行在当前指纹下全为绿；愿望完整——每一
              W 行都在某要求的区间之内或被标注
              not-a-requirement——；每一现存要求都被某个带行
              的现存行为的 covers 点名；并且没有未覆盖的源
              文件；声明的 artifacts 是文档签署的产出。语料
              的大小由愿望的大小决定——而非口味。任何不足都
              意味着「尚未」——report.md 点名缺口。

law:your-tree 行与构建在你的工作树中、从其根目录、仅带 PATH 与
              HOME，外加工作树自己的 ENV 文件（若存在；每行一条
              KEY=VALUE，按原样导出）运行——不传其它任何东西。
              项目所需之物以显式相对路径调用（./tool、bin/helper），
              并把此类文件放进该行为的 scope。构建一经
              声明便最先运行；其失败使本次运行的行全部停止。
              在命令内部自行清理：遗留的文件会改变树。

SCOPE——行依赖什么。每个行为的 scope 以从根起的通配列出其文件
（「**」跨越目录）。一行重新执行，当：其字节变了、其行为 scope
的文件变了、环境变了、或它从未执行过。scope 未变——该行永不再
执行。有行而无 scope 的行为每次运行都重跑：行必须始终有 scope。
不被任何 scope 覆盖的文件在 state.md 中分列：文档 artifacts
声明的立在「artifacts (declared)」之下，其余是未覆盖源文件——
用 scope 覆盖它们或加以声明，否则它们挡住裁决。

拒绝。被拒的提交留在 inbox.yaml——文档的字节一概不写；机器
把拒绝本身作为自己的事实记入日志，state.md 对其计数。
report.md 点名字段、事实与法律 id。在同一文件里修好被点名
之处再运行——拒绝的代价是一次运行，毫无损失。红行按原因分组列于 report.md。

文档的形状（复制 template.yaml 并编辑）：

  notes: [{text: 自由文本}]
  build: "./build all"
  remove: [old/thing, R2]
  requirements:
    - id: R1                 # R 后接字母、数字与点
      brief: [W1..W4, W9]    # 该要求占据的愿望行
  not-a-requirement:
    - lines: [W5]
      quote: "标题"           # 所标各行自己的文字，逐字
      reason: 标题，不是要求
  artifacts: [out/**]        # 工作声明的产出
  behaviors:
    - id: tool/build         # [a-z0-9-] 段，以 / 相连
      name: 一行
      prose: 用户得到什么，为何如此
      scope: [src/**]        # 该行为的文件，从工作根起的通配
      covers: [R1]           # 该行为证明的要求
      rows:
        - run: "./tool src out"
          expect: {rc: 0, out: "contains: ok",
                   files: {out/built.txt: "contains: done"}}

expect 的每个字段都可省略，但 rc、out、files 至少要有其一
——空的 expect 被拒绝；rc 缺省即要求 0；带「contains: 」前缀
的值匹配子串，不带则匹配精确字节。run 是一条命令：shell 行
或 argv 列表。W 区间为 W<n> 或 W<a>..W<b>，须递增，以逗号
相连——它们只点名愿望的编号行。

区文件。contract.md（本页；contract-ru.md、contract-zh.md 为其
镜像）template.yaml | wish.md inbox.yaml——由你写入 |
corpus.yaml 现行语料 | state.md 状态、漂移、简报地图、下一步 |
report.md 最近运行或拒绝 | knowledge.md 笔记与教训 | spec/
每个行为一页 | journal/ 正典——永不手改 | cache/ tmp/ 机器
自有。一条记录立于区外：工作根的 AGENTS.md 载有机器的标记
块（其发现钩子），每次运行都将其保持完整；块外的字节永不
触碰。
