# BuildSift

**构建失败了？BuildSift 帮你找出真正有用的那一条错误。**

不用 AI，不上传日志，只基于本机日志给出可核验的根因。

[English](README.md)

![BuildSift 把嘈杂构建日志压缩为一个根因、证据和下一步检查](docs/hero.svg)

## 为什么需要 BuildSift？

构建工具很擅长生成日志，却常常不擅长说明是哪一行最先引发了失败。一个文件缺失，可能最终变成数千行堆栈、跳过模块和次生错误。

BuildSift 会压缩这些噪声：对具体失败信号排序、折叠后续级联错误，并回指原始证据行，让每个判断都能人工核验。

### 处理前

```text
[INFO] Reactor Summary for payments-parent 1.4.0:
[INFO] payments-api ................................ FAILURE
[INFO] payments-service ............................ SKIPPED
[INFO] payments-web ................................ SKIPPED
[ERROR] Failed to execute goal org.apache.maven.plugins:maven-pmd-plugin:...
... 后面还有 2,184 行 ...
[ERROR] Re-run Maven using the -X switch to enable full debug logging.
```

### 处理后

```text
BUILDSIFT DIAGNOSIS
===================
ROOT CAUSE [HIGH · MAVEN]
NoSuchFileException: ~/work/quality/target/pmd.xml

EVIDENCE
  L1842  NoSuchFileException: ~/work/quality/target/pmd.xml

CASCADE
  17 additional failure signals folded

NEXT CHECK
  mvn -e -X
```

以上仅为展示格式；BuildSift 的结论始终来自你的实际日志。

## 30 秒快速开始

使用 Go 安装：

```bash
go install github.com/topchen2025/buildsift/cmd/buildsift@latest
```

也可以从 [Releases](https://github.com/topchen2025/buildsift/releases/latest) 下载对应平台的可执行文件并放入 `PATH`。然后包装原命令：

```bash
buildsift -- mvn test
```

BuildSift 会正常流式显示原命令输出。命令失败时，它在末尾打印简短诊断，并保留原命令的退出状态。

如果已经检出源码，可以改用：

```bash
go install ./cmd/buildsift
```

## 其他输入方式

分析已有日志文件：

```bash
buildsift build.log
```

直接分析 CI 输出：

```bash
gh run view --log-failed | buildsift
```

BuildSift v0.1 聚焦 Maven、Gradle、npm/pnpm 和 Docker/Compose 构建失败。

## 为什么使用确定性规则？

构建日志是证据，不是创作素材。

- **可解释：** 每个结论都回指产生它的原始日志行。
- **可复现：** 相同日志和规则集得到相同结果。
- **速度快：** 不下载模型，不等待 API，也不需要账号。
- **承认不确定：** 置信度不足时返回“无法判断”，不会强行猜测。
- **容易测试：** 每条规则都能配一份脱敏日志和预期结果。

## 隐私优先

BuildSift 只在本地分析日志，不会把日志发送到服务器。分析器不需要 API Key，也不会发起网络请求；被包装的构建命令仍可像原来一样访问网络。

诊断证据会遮盖常见的 token、密码、URL 凭据和用户主目录模式。BuildSift 包装命令时，原始流式输出不会被修改。自动脱敏属于尽力而为，并非绝对保证，分享前仍需人工检查。

日志仍可能包含凭据、私有路径、源码片段或客户数据。把日志附到 Issue 或分享给他人之前，请先检查并脱敏。安全问题报告方式见 [SECURITY.md](SECURITY.md)。

## 设计原则

1. 找最早出现的具体根因，而不是声音最大的最终报错。
2. 证据优先，不做无依据推测。
3. 折叠后果，但不隐藏原始日志。
4. 保留被包装命令的输出和退出状态。
5. 默认本地运行、速度快、无额外依赖。

## 路线图

- 扩充 Maven、Gradle、npm/pnpm、Docker/Compose 的高质量测试样本。
- 增加 JSON 和 Markdown 输出，支持 CI 注解与 Issue 报告。
- 发布官方 GitHub Action。
- 使用对抗性样本持续扩大凭据与私有路径的脱敏覆盖范围。
- 扩展社区维护的规则集，同时避免把核心过早做成复杂插件框架。

路线图会刻意保持克制：BuildSift 应该先变得更准确，再变得更可配置。

## 参与贡献

最有价值的贡献，是一份经过脱敏的真实失败日志，以及你已经确认的根因。它能把一次令人头疼的问题变成长期保护所有用户的回归测试。

开发流程和样本要求见 [CONTRIBUTING.md](CONTRIBUTING.md)。安全漏洞请通过 GitHub Security Advisories 私下报告，不要提交公开 Issue。

## 许可证

BuildSift 使用 [MIT License](LICENSE)。
