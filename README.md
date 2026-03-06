# CC-Start

Claude Code 启动器 - 快速切换不同 API 供应商。

## 安装

```bash
go install github.com/wujunwei/cc-start@latest
```

## 使用

### 首次配置

```bash
cc-start setup
```

### 启动 Claude Code

```bash
# 使用默认配置启动
cc-start

# 使用指定配置启动
cc-start moonshot

# 传递参数给 claude
cc-start -- --dangerously-skip-permissions
```

### 配置管理

```bash
# 列出所有配置
cc-start list

# 设置默认配置
cc-start default moonshot

# 删除配置
cc-start delete moonshot
```

## 支持的供应商

| 供应商 | Base URL | 默认模型 |
|--------|----------|----------|
| Anthropic | https://api.anthropic.com | claude-sonnet-4-5-20250929 |
| Moonshot | https://api.kimi.com/coding/ | moonshot-v1-8k |
| BigModel | https://open.bigmodel.cn/api/anthropic | glm-5 |
| DeepSeek | https://api.deepseek.com/anthropic | deepseek-chat |

## 配置文件

配置存储在 `~/.cc-start/profiles.json`
