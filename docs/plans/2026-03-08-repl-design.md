# REPL 交互模式设计文档

## 概述

为 cc-start 添加交互式 REPL 模式，用于管理 API 供应商配置。

## 需求总结

| 项目 | 选择 |
|------|------|
| REPL 类型 | 配置管理型 |
| 界面风格 | 命令行 + 表格美化 + 颜色 |
| 命令集 | 完整集（15+ 命令） |
| 历史补全 | 完整支持（持久化 + 智能提示） |
| API 测试 | 最小验证（成功/失败） |
| 启动方式 | 无参数进入 REPL |
| 配置编辑 | 交互式向导（复用 TUI） |

## 架构设计

```
┌─────────────────────────────────────────────────┐
│                   cc-start                       │
├─────────────────────────────────────────────────┤
│  cmd/repl.go          - REPL 入口命令           │
│  internal/repl/       - REPL 核心模块           │
│    ├── repl.go        - REPL 主循环             │
│    ├── parser.go      - 命令解析器              │
│    ├── commands.go    - 命令注册与执行          │
│    ├── completer.go   - Tab 补全 + 智能提示     │
│    ├── history.go     - 历史管理（持久化）      │
│    └── ui.go          - 输出格式化（表格/颜色） │
│  internal/config/     - 现有配置模块（复用）    │
│  internal/tui/setup/  - 现有 TUI 组件（复用）   │
└─────────────────────────────────────────────────┘
```

## 命令列表

| 命令 | 别名 | 说明 |
|------|------|------|
| `list` | `ls` | 表格展示所有配置 |
| `use <name>` | `switch` | 切换当前会话配置 |
| `current` | `status` | 显示当前使用的配置 |
| `default <name>` | - | 设置默认配置 |
| `show <name>` | - | 显示配置详情 |
| `add` | `new` | 启动交互向导添加配置 |
| `edit <name>` | - | 启动向导编辑配置 |
| `delete <name>` | `rm` | 删除配置（需确认） |
| `copy <from> <to>` | `cp` | 复制配置 |
| `rename <old> <new>` | `mv` | 重命名配置 |
| `test <name>` | - | 测试 API 连通性 |
| `export [file]` | - | 导出配置（默认 stdout） |
| `import <file>` | - | 从文件导入配置 |
| `history` | - | 显示命令历史 |
| `help [cmd]` | `?` | 显示帮助 |
| `clear` | `cls` | 清屏 |
| `exit` | `quit`, `q` | 退出 REPL |
| `run [profile] [-- args]` | - | 启动 claude |

**会话配置 vs 默认配置：**
- `use` 只影响当前 REPL 会话
- `default` 持久化到配置文件

## 输出格式化

### `list` 输出示例

```
┌──────────┬─────────────────────────────┬──────────────────┬─────────┐
│  Name    │  Base URL                   │  Model           │  Status │
├──────────┼─────────────────────────────┼──────────────────┼─────────┤
│  anthropic│ https://api.anthropic.com  │ claude-sonnet-4-5│ default │
│  moonshot │ https://api.kimi.com/...   │ moonshot-v1-8k   │ current │
│  deepseek │ https://api.deepseek.com/..│ deepseek-chat    │         │
└──────────┴─────────────────────────────┴──────────────────┴─────────┘
```

### 颜色规范

- 🟢 绿色：成功、默认配置
- 🔵 蓝色：当前使用
- 🟡 黄色：警告、待确认
- 🔴 红色：错误、失败
- ⚪ 灰色：提示信息

### 状态标识

- `default` - 默认配置
- `current` - 当前会话使用
- 可组合：`default/current`

### 交互提示示例

```
cc-start [moonshot]> use deepseek
✓ 已切换到 deepseek

cc-start [deepseek]> test
● 测试中... ✓ 连接成功
```

## 历史与补全

### 历史文件

- 存储位置：`~/.cc-start/history`
- 最大条数：1000 条（FIFO 淘汰）
- 格式：每行一条命令

### 快捷键

| 快捷键 | 功能 |
|--------|------|
| `↑` / `↓` | 遍历历史命令 |
| `Ctrl+R` | 搜索历史 |
| `Tab` | 补全命令/配置名 |
| `Shift+Tab` | 反向补全 |
| `Ctrl+L` | 清屏 |
| `Ctrl+C` | 取消当前输入 |
| `Ctrl+D` | 退出 REPL |

### 补全逻辑

```
# 命令补全
li<Tab> → list
u<Tab>  → use

# 配置名补全（根据上下文）
use mo<Tab> → use moonshot
delete d<Tab> → delete deepseek

# 参数提示
test <Tab> → 显示所有配置名
export <Tab> → 显示文件路径（可选）
```

### 智能提示

```
cc-start> us
  use <name>    切换当前配置
  user          (无匹配)
```

## 启动流程

### 执行流程

```
cc-start 执行流程：
┌─────────────────────────────────────────────┐
│  有子命令？ ─────────────────→ 执行子命令    │
│     ↓ 否                                    │
│  进入 REPL 模式                             │
└─────────────────────────────────────────────┘
```

### 命令行为

| 命令 | 行为 |
|------|------|
| `cc-start` | 进入 REPL |
| `cc-start run [profile] [-- args]` | 启动 claude |
| `cc-start list` | 列出配置 |
| `cc-start default <name>` | 设置默认 |
| `cc-start delete <name>` | 删除配置 |
| `cc-start setup` | 配置向导 |
| `cc-start version` | 版本信息 |

### REPL 内启动 claude

```
cc-start> run                # 用当前配置启动 claude
cc-start> run moonshot       # 用指定配置启动
cc-start> run -- --help      # 传参给 claude
```

## 技术选型

### 新增依赖

| 库 | 用途 | 说明 |
|---|------|------|
| `github.com/c-bata/go-prompt` | REPL 输入体验 | 历史、补全、搜索 |
| `github.com/olekukonko/tablewriter` | 表格输出 | `list` 命令美化 |
| `github.com/fatih/color` | 颜色输出 | 状态高亮 |

### 复用现有模块

- `internal/config` - 配置管理
- `internal/tui/setup` - 添加/编辑向导
- `internal/launcher` - 启动 claude

### 文件结构

```
internal/repl/
├── repl.go        # REPL 主循环、go-prompt 集成
├── commands.go    # 所有命令实现
├── completer.go   # Tab 补全逻辑
└── ui.go          # 表格、颜色输出
```

### 历史文件处理

- 自行实现简单的文件读写（无需额外依赖）
- 路径：`~/.cc-start/history`

## 错误处理

### 配置文件不存在

```
$ cc-start
⚠ 尚未配置任何供应商
运行 'setup' 创建配置，或 'help' 查看帮助

cc-start>
```

### 配置名为空或不存在

```
cc-start> use
✗ 请指定配置名

cc-start> use unknown
✗ 配置 'unknown' 不存在，可用：anthropic, moonshot, deepseek
```

### 删除确认

```
cc-start> delete moonshot
确认删除配置 'moonshot'？  不可恢复
```

### API 测试失败

```
cc-start> test moonshot
● 测试中... ✗ 连接失败: 网络超时
```

### 导入冲突处理

```
cc-start> import backup.json
发现 2 个同名配置：
  - moonshot (已有)
  - deepseek (已有)
覆盖(O) / 跳过(S) / 重命名(R) / 取消(C)？
```

### Ctrl+C 中断

- 输入中途：清空当前行
- 向导中途：退出并提示"已取消"

## 测试策略

### 单元测试

```
internal/repl/
├── commands_test.go    # 命令解析与执行
├── completer_test.go   # 补全逻辑
└── parser_test.go      # 输入解析
```

### 测试覆盖点

| 模块 | 测试项 |
|------|--------|
| 命令解析 | 命令名、别名、参数提取 |
| 补全 | 命令补全、配置名补全、上下文感知 |
| 配置操作 | 增删改查、默认设置、复制重命名 |
| 导入导出 | JSON 格式、冲突处理 |
| 历史记录 | 持久化、最大条数限制 |

### 集成测试

- 启动流程：无参数 → REPL，`run` → 启动 claude
- 端到端：添加配置 → 切换 → 测试 → 删除

### Mock 策略

- `test` 命令：Mock HTTP 响应，避免真实 API 调用
- `run` 命令：Mock `launcher.Launch`，验证参数传递
