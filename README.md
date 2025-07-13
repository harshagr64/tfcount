# tfcount 🚀

A CLI tool to summarize Terraform plan outputs by resource type and action (create, update, delete).
> ⚠️ **Note:** `tfcount` is currently supported **only on macOS** (Apple Silicon). Support for Linux and Windows coming soon!

## Features ✨

- Parses Terraform plan output (JSON) and summarizes actions by resource type:
  - **Create**
  - **Update**
  - **Delete**
- Simple command-line interface
- Helps quickly understand the impact of Terraform changes

## Requirements 📋

- [Go](https://golang.org/) 1.18 or higher
- Terraform (for generating plan outputs)

## Installation ⚙️
> **Platform Support**: This version supports **macOS only**. Other platforms are not yet supported.
>
> 🛠️ **No Go Required**: You do **not** need Go installed for the quick install method.

### Quick Install Command 🚀

```bash
curl -sSL https://gist.githubusercontent.com/harshagr64/a105164f646492ad99346bddb5ff107b/raw/a2c0afe169dd13ede5f827ac002f2c9ffcf8bddb/install-tfcount.sh | bash
tfcount help
```

### Install from Source 🛠️

```bash
git clone https://github.com/harshagr64/tfcount.git
cd tfcount
go build -o tfcount
sudo mv tfcount /usr/local/bin/
tfcount help
```

## Usage 🛠️

Run the CLI tool:

```bash
tfcount plan
```

Use the `--terragrunt` flag to run with Terragrunt:

```bash
tfcount plan --terragrunt
```

### Options 🏷️

- `-h, --help`: Show help message.

## How It Works 🧐

1. Runs `terraform plan` (or `terragrunt plan`) and saves the output to a file.
2. Runs `terraform show -json` (or `terragrunt show -json`) to parse the plan output.
3. Summarizes the changes by resource type and action (create, update, delete).
4. Displays a summary of the changes in a user-friendly format.

## Example Output 📊

```plaintext
📊 Resource Change Summary:
aws_instance:
    + create: 2
    ~ update: 1
    - delete: 1
aws_s3_bucket:
    + create: 1
```

---

Feel free to suggest improvements or report issues! 💡
