# BuildSift

**把失败构建日志变成紧凑、带原始行号的证据包，交给 Claude、Codex、CI，也交给人。**

本地运行、结果确定、不上传日志。BuildSift 保留有用的失败信号，折叠级联报错，并指回准确的原始行。

[English](README.md)

![BuildSift 把嘈杂构建日志压缩为一个根因、证据和下一步检查](docs/hero.svg)

## 30 秒开始试用

从 [Releases](https://github.com/topchen2025/buildsift/releases/latest) 下载最新版，或使用 Go 安装：

```bash
go install github.com/topchen2025/buildsift/cmd/buildsift@latest
```

分析已有日志：

```bash
buildsift build.log
```

也可以包装原命令，同时保留它的原始输出和退出状态：

```bash
buildsift -- mvn test
```

BuildSift v0.1 聚焦 Maven、Gradle、npm/pnpm 和 Docker/Compose 构建失败。

## 从日志噪声到证据包

```text
BUILDSIFT DIAGNOSIS
===================
ROOT CAUSE [HIGH · MAVEN]
NoSuchFileException: ~/work/quality/target/pmd.xml

EVIDENCE
  L1842  NoSuchFileException: ~/work/quality/target/pmd.xml

CASCADE
  17 additional failure signal(s) folded

NEXT CHECK
  mvn -e -X
```

以上仅为格式示例。BuildSift 只报告输入日志中实际存在的证据；没有命中受支持的模式时，不会编造解释。

结果足够短，可以直接交给 Claude 或 Codex；结构足够稳定，可以用于 CI；同时保留明确行号，方便人工回查原日志。

## 配合 Agent 和自动化使用

生成便于分享的文本证据包：

```bash
buildsift build.log > evidence.txt
```

为脚本和 CI 生成机器可读结果：

```bash
buildsift --json build.log
```

也可以分析 GitHub Actions 运行中的失败日志：

```bash
gh run view --log-failed | buildsift
```

需要仓库内集成时，可使用 [BuildSift GitHub Action](docs/github-action.md)，并固定版本：

```yaml
- name: Analyze failed build log
  uses: topchen2025/buildsift@v0.1.0
  with:
    log-path: build.log
```

## 为什么使用确定性规则？

构建日志是证据，不是创作素材。

- **可核验：** 证据保留原始日志行号。
- **可复现：** 相同日志和规则集产生相同结果。
- **默认保护隐私：** 分析在本地完成，不上传日志。
- **容易采用：** 不需要模型、API Key、账号或在线服务。
- **承认不确定：** 无法识别的失败会返回“无法判断”，而不是猜测。

BuildSift 不替代 AI 工具；它为 AI 提供更小、更有依据的输入，避免让 Agent 在整份嘈杂日志中盲目搜索。

## 能做什么，不能做什么

BuildSift 会对具体失败信号排序，优先选择最早出现的可执行原因，折叠已识别的后续错误，并输出证据与下一步检查。输入可以是文件、标准输入，也可以是被包装命令的输出。

它不是通用调试器，目前的规则集也无法识别所有构建失败。公开样本与评估方法见 [docs/benchmark.md](docs/benchmark.md)；这些结果不代表通用准确率或固定压缩率。

## 隐私与脱敏

分析器不会发起网络请求，也不会把日志发送到服务器；被包装的构建命令仍可像原来一样访问网络。

诊断证据会遮盖常见的 token、密码、URL 凭据和用户主目录模式。自动脱敏属于尽力而为，并非绝对保证。日志仍可能包含凭据、私有路径、源码片段或客户数据，分享前请人工检查。详情见 [SECURITY.md](SECURITY.md)。

## 参与贡献

最有价值的贡献，是一份经过脱敏的真实失败日志，以及已经确认的根因。它能把一次故障变成帮助所有人的回归样本。

开发流程和样本要求见 [CONTRIBUTING.md](CONTRIBUTING.md)。安全漏洞请通过 GitHub Security Advisories 私下报告，不要提交公开 Issue。

## 许可证

[MIT](LICENSE)
