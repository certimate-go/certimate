<h1 align="center">🔒 Certimate</h1>

<div align="center">

[![Stars](https://img.shields.io/github/stars/certimate-go/certimate?style=flat)](https://github.com/certimate-go/certimate)
[![Forks](https://img.shields.io/github/forks/certimate-go/certimate?style=flat)](https://github.com/certimate-go/certimate)
[![Docker Pulls](https://img.shields.io/docker/pulls/certimate/certimate?style=flat)](https://hub.docker.com/r/certimate/certimate)
[![Release](https://img.shields.io/github/v/release/certimate-go/certimate?style=flat&sort=semver)](https://github.com/certimate-go/certimate/releases)
[![License](https://img.shields.io/github/license/certimate-go/certimate?style=flat)](https://mit-license.org/)

</div>

<div align="center">

中文 ｜ [English](README_EN.md)

</div>

---

## 🚩 项目简介

做个人产品或者在中小企业里负责运维的同学，会遇到要管理多个域名的情况，需要给域名申请证书。但是手动申请证书有以下缺点：

- 😱 麻烦：申请证书并部署到服务的流程虽不复杂，但也挺麻烦的，尤其是你有多个域名需要维护的时候。
- 😭 易忘：另外当前免费证书的有效期只有 90 天，这就要求你定期的操作，增加了工作量的同时，你也很容易忘掉续期，从而导致网站访问不了。

Certimate 就是为了解决上述问题而产生的，它具有以下优势：

- **本地部署**：一键安装，只需要下载二进制文件，然后直接运行即可。同时也支持 Docker 部署、源代码部署等方式。​
- **数据安全**：由于是私有部署，所有数据均存储在自己的服务器上，不会经过第三方，确保数据的隐私和安全。​
- **操作简单**：简单配置即可轻松申请 SSL 证书并部署到指定的目标上，在证书即将过期前自动续期，从申请证书到使用证书完全自动化，无需人工操作。​

Certimate 旨在为用户提供一个安全、简便的 SSL 证书管理解决方案。

## 💡 功能特性

- 灵活的工作流编排方式，证书从申请到部署完全自动化；
- 支持单域名、多域名、泛域名证书，可选 RSA、ECC 签名算法；
- 支持 DNS-01（即基于域名解析验证）、HTTP-01（即基于文件验证）两种质询方式；
- 支持 PEM、PFX、JKS 等多种格式输出证书；
- 支持 60+ 域名托管商（如阿里云、腾讯云、Cloudflare 等，[点此查看完整清单](https://docs.certimate.me/docs/reference/providers#supported-dns-providers)）；
- 支持 110+ 部署目标（如 Kubernetes、CDN、WAF、负载均衡等，[点此查看完整清单](https://docs.certimate.me/docs/reference/providers#supported-hosting-providers)）；
- 支持邮件、钉钉、飞书、企业微信、Webhook 等多种通知渠道；
- 支持 Let's Encrypt、Actalis、Google Trust Services、SSL.com、ZeroSSL 等多种 ACME 证书颁发机构；
- 更多特性等待探索。

## ⏱️ 快速启动

**1 分钟部署 Certimate！**

以二进制部署为例，从 [GitHub Releases](https://github.com/certimate-go/certimate/releases) 页面下载预先编译好的二进制可执行文件压缩包，解压缩后在终端中执行：

```bash
./certimate serve
```

浏览器中访问 `http://127.0.0.1:8090`。

初始的管理员账号及密码：

- 账号：`admin@certimate.fun`
- 密码：`1234567890`

即刻使用 Certimate。

如何使用 Docker 或其他部署方式请参考文档。

## 📄 使用手册

请访问文档站 [docs.certimate.me](https://docs.certimate.me/) 以阅读使用手册。

相关文章：

- [《升级指南：迁移到 v0.4》](https://docs.certimate.me/docs/migrations/migrate-to-v0.4)
- [《使用 CNAME 完成 ACME DNS-01 质询》](https://docs.certimate.me/blog/cname)
- [《Why Certimate?》](https://docs.certimate.me/blog/why-certimate)

## ⭐ 运行界面

[![Screenshot](https://i.imgur.com/4DAUKEE.gif)](https://www.bilibili.com/video/BV1xockeZEm2)

## 🤝 参与贡献

Certimate 是一个免费且开源的项目。我们欢迎任何人为 Certimate 做出贡献，以帮助改善 Certimate。包括但不限于：提交代码、反馈缺陷、交流想法，或分享你基于 Certimate 的使用案例。同时，我们也欢迎用户在个人博客或社交媒体上分享 Certimate。

如果你想要贡献代码，请先阅读我们的[贡献指南](./CONTRIBUTING.md)。

请在 https://github.com/certimate-go/certimate 提交 [Issues](https://github.com/certimate-go/certimate/issues) 和 [Pull Requests](https://github.com/certimate-go/certimate/pulls)。

#### 感谢以下贡献者对 Certimate 做出的贡献：

[![Contributors](https://contrib.rocks/image?repo=certimate-go/certimate)](https://github.com/certimate-go/certimate/graphs/contributors)

## ⛔ 免责声明

Certimate 遵循 [MIT License](https://opensource.org/licenses/MIT) 开源协议，完全免费提供，旨在“按现状”供用户使用。作者及贡献者不对使用本软件所产生的任何直接或间接后果承担责任，包括但不限于性能下降、数据丢失、服务中断、或任何其他类型的损害。

**无任何保证**：本软件不提供任何明示或暗示的保证，包括但不限于对特定用途的适用性、无侵权性、商用性及可靠性的保证。

**用户责任**：使用本软件即表示您理解并同意承担由此产生的一切风险及责任。

## 🌐 加入社群

- [Telegram](https://t.me/+ZXphsppxUg41YmVl)
- 微信群聊（超 200 人需邀请入群，可先加作者好友）

  <img src="https://i.imgur.com/8xwsLTA.png" width="200"/>

## 🚀 Star 趋势图

[![Stargazers over time](https://starchart.cc/certimate-go/certimate.svg?variant=adaptive)](https://starchart.cc/certimate-go/certimate)
