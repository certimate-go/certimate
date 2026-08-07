# Contribution Guide

<div align="center">

English ｜ [简体中文](CONTRIBUTING_zh.md)

</div>

Thank you for taking the time to improve Certimate! Below is a guide for submitting a PR (Pull Request) to the Certimate repository.

We need to be nimble and ship fast given where we are, but we also want to make sure that contributors like you get as smooth an experience at contributing as possible. We've assembled this contribution guide for that purpose, aiming at getting you familiarized with the codebase & how we work with contributors, so you could quickly jump to the fun part.

Index:

- [Development](#development)
  - [Prerequisites](#prerequisites)
  - [Backend Code](#backend-code)
  - [Frontend Code](#frontend-code)
- [Submitting PR](#submitting-pr)
  - [Pull Request Process](#pull-request-process)
- [Getting Help](#getting-help)

---

## Development

> [!IMPORTANT]
>
> This project uses `git submodule`. When cloning, please ensure that the submodule is also pulled.
>
> ```bash
> # Option 1: Direct recursive pull during cloning
> git clone --recurse-submodules https://github.com/certimate-go/certimate.git
>
> # Option 2: Clone first, then initialize
> git clone https://github.com/certimate-go/certimate.git
> cd certimate
> git submodule update --init --recursive
> ```

### Prerequisites

- Go 1.25+ (for backend code changes)
- Node.js 22.12+ (for frontend code changes)

### Backend Code

The backend code of Certimate is developed using Golang. It is a monolithic application based on [Pocketbase](https://github.com/pocketbase/pocketbase).

Once you have made changes to the backend code in Certimate, follow these steps to run the project:

1. Navigate to the root directory.
2. Install dependencies:
   ```bash
   go mod vendor
   ```
3. Start the local development server:
   ```bash
   go run main.go serve
   ```

This will start a web server at `http://localhost:8090` using the prebuilt WebUI located in `/ui/dist`.

> If you encounter an error `ui/embed.go: pattern all:dist: no matching files found`, please refer to _[Frontend Code](#frontend-code)_ and build WebUI first.

**Before submitting a PR to the main repository, you should:**

- Format your source code by using [gofumpt](https://github.com/mvdan/gofumpt). Recommended using VSCode and installing the gofumpt plugin to automatically format when saving.
- Adding unit or integration tests for your changes (with go standard library `testing` package).

### Frontend Code

The frontend code of Certimate is developed using TypeScript. It is a SPA based on [React](https://github.com/facebook/react) and [Vite](https://github.com/vitejs/vite).

Once you have made changes to the backend code in Certimate, follow these steps to run the project:

1. Navigate to the `/ui` directory.
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the local development server:
   ```bash
   npm run dev
   ```

This will start a web server at `http://localhost:5173`. You can now access the WebUI in your browser.

After completing your changes, build the WebUI so it can be embedded into the Go package:

```bash
npm run build
```

**Before submitting a PR to the main repository, you should:**

- Format your source code by using [ESLint](https://github.com/eslint/eslint). Recommended using VSCode and installing the ESLint plugin to automatically format when saving.

## Submitting PR

Before opening a Pull Request, please open an issue to discuss the change and get feedback from the maintainers. This will helps us:

- To understand the context of the change.
- To ensure it fits into Certimate's roadmap.
- To prevent us from duplicating work.
- To prevent you from spending time on a change that we may not be able to accept.

### Pull Request Process

1. Fork the repository, and then checkout `main` branch.
2. Before you draft a PR, please open an issue to discuss the changes you want to make.
3. Create a new branch for your changes.
4. Please add tests for your changes accordingly.
5. Ensure your code passes the existing tests.
6. Please link the issue in the PR description.
7. Get merged!

> [!IMPORTANT]
>
> It is recommended to create a new branch from `main` for each bug fix or feature. If you plan to submit multiple PRs, ensure the changes are in separate branches for easier review and eventual merge.
>
> Keep each PR focused on a single feature or fix.

## Getting Help

If you ever get stuck or get a burning question while contributing, simply shoot your queries our way via the GitHub issues.
