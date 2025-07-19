# 贡献指南

非常感谢你抽出时间来帮助改进 Certimate！以下是向 Certimate 提交 Pull Request 时的操作指南。

我们需要保持敏捷和快速迭代，同时也希望确保贡献者能获得尽可能流畅的参与体验。这份贡献指南旨在帮助你熟悉代码库和我们的工作方式，让你可以尽快进入有趣的开发环节。

索引：

- [开发](#开发)
  - [要求](#要求)
  - [后端代码](#后端代码)
  - [前端代码](#前端代码)
- [提交 PR](#提交-pr)
  - [提交流程](#提交流程)
- [获取帮助](#获取帮助)

---

## 开发

### 要求

- Go 1.24+（用于修改后端代码）
- Node.js 22.0+（用于修改前端代码）

### 后端代码

Certimate 的后端代码是使用 Golang 开发的，是一个基于 [Pocketbase](https://github.com/pocketbase/pocketbase) 构建的单体应用。

假设你已经对 Certimate 的后端代码做出了一些修改，现在你想要运行它，请遵循以下步骤：

1. 进入根目录；
2. 安装依赖：
   ```bash
   go mod vendor
   ```
3. 启动本地开发服务：
   ```bash
   go run main.go serve
   ```

这将启动一个 Web 服务器，默认运行在 `http://localhost:8090`，并使用来自 `/ui/dist` 的预构建管理页面。

> 如果你遇到报错 `ui/embed.go: pattern all:dist: no matching files found`，请参考“[前端代码](#前端代码)”这一小节构建 WebUI。

**在向主仓库提交 PR 之前，你应该：**

- 使用 [gofumpt](https://github.com/mvdan/gofumpt) 格式化你的代码。推荐使用 VSCode，并安装 gofumpt 插件，以便在保存时自动格式化。
- 为你的改动添加单元测试或集成测试（使用 Go 标准库中的 `testing` 包）。

### 前端代码

Certimate 的前端代码是使用 TypeScript 开发的，是一个基于 [React](https://github.com/facebook/react) 和 [Vite](https://github.com/vitejs/vite) 构建的单页应用。

假设你已经对 Certimate 的前端代码做出了一些修改，现在你想要运行它，请遵循以下步骤：

1. 进入 `/ui` 目录；
2. 安装依赖：
   ```bash
   npm install
   ```
3. 启动 Vite 开发服务器：
   ```bash
   npm run dev
   ```

这将启动一个 Web 服务器，默认运行在 `http://localhost:5173`，你可以通过浏览器访问来查看运行中的 WebUI。

完成修改后，运行以下命令来构建 WebUI，以便它能被嵌入到 Go 包中：

```bash
npm run build
```

**在向主仓库提交 PR 之前，你应该：**

- 使用 [ESLint](https://github.com/eslint/eslint) 格式化你的代码。推荐使用 VSCode，并安装 ESLint 插件，以便在保存时自动格式化。

## 提交 PR

在提交 PR 之前，请先创建一个 Issue 来讨论你的修改方案，并等待来自项目维护者的反馈。这样做有助于：

- 让我们充分理解你的修改内容；
- 评估修改是否符合项目路线图；
- 避免重复工作；
- 防止你投入时间到可能无法被合并的修改中。

### 提交流程

1. Fork 本仓库；
2. 在提交 PR 之前，请先发起 Issue 讨论你想要做的修改；
3. 为你的修改创建一个新的分支；
4. 请为你的修改添加相应的测试；
5. 确保你的代码能通过现有的测试；
6. 请在 PR 描述中关联相关 Issue；
7. 等待合并！

> [!IMPORTANT]
>
> 建议为每个新功能或 Bug 修复创建一个从 `main` 分支派生的新分支。如果你计划提交多个 PR，请保持不同的改动在独立分支中，以便更容易进行代码审查并最终合并。
>
> 保持一个 PR 只实现一个功能或修复。

## 获取帮助

如果你在贡献过程中遇到困难或问题，可以通过 GitHub Issues 向我们提问。
