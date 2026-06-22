<h1 align="center">🔒 Certimate</h1>

<div align="center">

[![Stars](https://img.shields.io/github/stars/certimate-go/certimate?style=flat)](https://github.com/certimate-go/certimate)
[![Forks](https://img.shields.io/github/forks/certimate-go/certimate?style=flat)](https://github.com/certimate-go/certimate)
[![Docker Pulls](https://img.shields.io/docker/pulls/certimate/certimate?style=flat)](https://hub.docker.com/r/certimate/certimate)
[![Release](https://img.shields.io/github/v/release/certimate-go/certimate?style=flat&sort=semver)](https://github.com/certimate-go/certimate/releases)
[![License](https://img.shields.io/github/license/certimate-go/certimate?style=flat)](https://mit-license.org/)

</div>

<div align="center">

[![Ask DeepWiki](https://deepwiki.com/badge.svg?style=flat)](https://deepwiki.com/certimate-go/certimate)
[![Ask Zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTUwMiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/certimate-go/certimate)

</div>

<div align="center">

English ｜ [简体中文](README_zh.md)

</div>

---

## 🚩 Introduction

An open-source and free self-hosted SSL certificates ACME tool, automates the full-cycle of issuance, deployment, renewal, and monitoring visually.

- **Self-hosted**: Private deployment. All data is stored locally, to ensure data privacy and security.
- **Cross Platforms**: Compatible with various operating systems, including Windows/Linux/macOS.
- **Zero Dependencies**: No need to install databases, runtimes, or any frameworks. Ready to use out of the box.
- **Low Resource Usage**: Extremely lightweight, requiring only ~16 MB of memory.
- **Easy to Use**: Automates certificate management with a visual workflow — all with just a few simple configurations.

## 💡 Features

- Flexible workflow orchestration, fully automation from certificate application to deployment.
- Supports requesting single/multiple/wildcard domain certificates, IP address certificates, with options for RSA or ECC key.
- Supports DNS-01 challenge and HTTP-01 challenge both.
- Supports various certificate formats such as PEM, PFX, JKS.
- Supports more than 70+ domain registrars (e.g., AWS, Cloudflare, GoDaddy, Alibaba Cloud, Tencent Cloud, etc. [Check out full providers](https://docs.certimate.me/en-US/docs/reference/providers#supported-dns-providers)).
- Supports more than 150+ deployment destinations (e.g., Kubernetes, CDN, WAF, load balancers, etc. [Check out full providers](https://docs.certimate.me/en-US/docs/reference/providers#supported-hosting-providers)).
- Supports multiple notification channels including email, Discord, Slack, Telegram, DingTalk, Feishu, WeCom, and more.
- Supports multiple ACME CAs including Let's Encrypt, Actalis, Google Trust Services, SSL.com, ZeroSSL, and more.
- More features waiting to be discovered.

## 🚀 Quick Start

**Run Certimate in 1 minute!**

<details>
<summary>👉 Binary Installation: </summary>

Download the archived package of precompiled executable files directly from [GitHub Releases](https://github.com/certimate-go/certimate/releases), extract and then execute:

```bash
./certimate serve
```

</details>

<details>
<summary>👉 Docker Installation: </summary>

```bash
docker run -d \
  --name certimate \
  --restart unless-stopped \
  -p 8090:8090 \
  -v /etc/localtime:/etc/localtime:ro \
  -v /etc/timezone:/etc/timezone:ro \
  -v $(pwd)/data:/app/pb_data \
  certimate/certimate:latest
```

</details>

Visit `http://127.0.0.1:8090` in your browser.

Default administrator account:

- Username: `admin@certimate.fun`
- Password: `1234567890`

Work with Certimate right now. Or read other content in the documentation to learn more.

## 📄 Documentation

For full documentation, please visit [docs.certimate.me](https://docs.certimate.me/).

Related articles:

> - [_Migrate to v0.4_](https://docs.certimate.me/en-US/docs/migrations/migrate-to-v0.4)
> - [_使用 CNAME 完成 ACME DNS-01 质询_](https://docs.certimate.me/en-US/blog/cname)
> - [_Why Certimate?_](https://docs.certimate.me/en-US/blog/why-certimate)

## 🖥️ Screenshot

[![Screenshot](https://i.imgur.com/4DAUKEE.gif)](https://www.youtube.com/watch?v=am_yzdfyNOE)

## 🤝 Contributing

Certimate is a free and open-source project, and your feedback and contributions are needed and always welcome. Contributions include but are not limited to: submitting code, reporting bugs, sharing ideas, or showcasing your use cases based on Certimate. We also encourage users to share Certimate on personal blogs or social media.

For those who'd like to contribute code, see our [Contribution Guide](./CONTRIBUTING.md).

[Issues](https://github.com/certimate-go/certimate/issues) and [Pull Requests](https://github.com/certimate-go/certimate/pulls) are opened at https://github.com/certimate-go/certimate.

#### Contributors

[![Contributors](https://contrib.rocks/image?repo=certimate-go/certimate)](https://github.com/certimate-go/certimate/graphs/contributors)

## ⛔ Disclaimer

This repository is available under the [MIT License](https://opensource.org/licenses/MIT), and distributed “as-is” without any warranty of any kind. The authors and contributors are not responsible for any damages or losses resulting from the use or inability to use this software, including but not limited to data loss, business interruption, or any other potential harm.

**No Warranties**: This software comes without any express or implied warranties, including but not limited to implied warranties of merchantability, fitness for a particular purpose, and non-infringement.

**User Responsibilities**: By using this software, you agree to take full responsibility for any outcomes resulting from its use.

## 🌐 Join the Community

- [Telegram](https://t.me/+ZXphsppxUg41YmVl)
- Wechat Group (contact to the author [@usual2970](https://github.com/usual2970) to getting invitation)

  <img src="https://i.imgur.com/8xwsLTA.png" width="200"/>

## ⭐ Star History

Star Certificate on GitHub and be instantly notified of new releases.

[![Stargazers over time](https://starchart.cc/certimate-go/certimate.svg?variant=adaptive)](https://starchart.cc/certimate-go/certimate)
