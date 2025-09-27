# Contributing to tfcount

Thank you for your interest in contributing to **tfcount**! 🎉  
We welcome all contributions.
Please follow these guidelines to help us review and accept your pull requests smoothly.

---

## 📋 Table of Contents

- [How to Contribute](#how-to-contribute)
- [Getting Started](#getting-started)
- [Development Guidelines](#development-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)
- [Community & Support](#community--support)

---

## How to Contribute

1. **Fork** this repository.
2. **Clone** your fork:  
   ```sh
   git clone https://github.com/<your-username>/tfcount.git
   ```
3. **Create a new branch** for your changes:  
   ```sh
   git checkout -b feature/your-feature-name
   ```
4. Make your changes and **commit** them.
5. **Push** to your fork and submit a **Pull Request**.

---

## Getting Started

This project is written in Go. Below are the steps to get the project running locally:
- Install [Go](https://golang.org/dl/) (version 1.18 or above recommended).
- Run `go mod tidy` to install dependencies.
- Build:  
  ```sh
  go build -o tfcount
  ```
- (Optional) Move to a location in your PATH so you can run it by name:
  ```sh
  mv tfcount /usr/local/bin/
  ```
- Run commands to verify functionality, e.g.:
  ```sh
  ./tfcount plan
  ./tfcount plan --terragrunt
  ```

---

## Development Guidelines

- **Write clear, concise code** following [Go best practices](https://github.com/golang/go/wiki/CodeReviewComments).
- Use `gofmt` or `go fmt` to format code before committing.
- Add **unit tests** for new features or bug fixes.
- Write **clear commit messages** (e.g., `fix: handle empty resource type`).
- Comment your code where necessary for clarity.

---

## Pull Request Process

- Ensure your branch is **up to date** with `main`.
- Pull Requests should be **atomic** and **focused** on a single topic.
- Reference the related issue in your PR description (if applicable).
- In **PR description**, include:
  - What problem it addresses
  - How you fixed it
  - Any new behavior or backward compatibility notes
  - Screenshots / output examples if applicable
  - Tag any relevant issues (e.g. “Fixes #45”)
- Label your PR appropriately (e.g. `bug`, `enhancement`).
- Be responsive to reviewer feedback.

---

## Reporting Issues

- Use the [Issue Tracker](https://github.com/harshagr64/tfcount/issues) to report bugs or request features.
- Provide as much detail as possible (logs, steps to reproduce, expected/actual behavior).

---

## Community & Support

- For questions, open a [Discussion](https://github.com/harshagr64/tfcount/discussions) or join the issue thread.
- Be kind and collaborative!

---

We’re really excited to have your help in improving **tfcount**! Thank you for being part of it 🎉

---
