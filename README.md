# goot: Your Powerful Terminal Task Manager 🚀

![Go](https://img.shields.io/badge/Go-GCD7DF?style=for-the-badge&logo=go&logoColor=white)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/zeerodex/goot?style=for-the-badge)
![License](https://img.shields.io/github/license/zeerodex/goot?style=for-the-badge)

**goot** is a fast, flexible, and modern task manager operating directly from your terminal. Built in Go, it offers a seamless experience through its command-line interface (CLI) and a rich terminal user interface (TUI). `goot` is designed not just for local task management but also includes a robust daemon and integrates with external task management services via its API capabilities, making it a central hub for your productivity.

## ✨ Features

* ⚙️ **Efficient CLI:** Perform quick actions like adding, completing, listing, and removing tasks with intuitive commands.
* 🖥️ **Interactive TUI:** A dynamic and visually appealing terminal user interface provides a comprehensive overview and interactive management of your tasks.
* 🤖 **Reliable Daemon:** A background daemon runs system-wide, enabling continuous operation and supporting advanced features.
* 🔌 **API Integration:** Connects with external task management services to synchronize and manage tasks across different platforms.
* 💾 **Local Persistence:** Your tasks are securely stored locally using sqlite, providing a dependable offline task management option.

## Get Started

### Prerequisites

Ensure you have Go installed (version 1.x or higher).

### Installation

1.  Clone the `goot` repository:
    ```bash
    git clone [https://github.com/zeerodex/goot.git](https://github.com/zeerodex/goot.git)
    ```
2.  Navigate into the project directory:
    ```bash
    cd goot
    ```
3.  Build the `goot` executable:
    ```bash
    go build -o goot
    ```
4.  *(Optional but Recommended)* Move `goot` to your PATH for easy access from anywhere:
    ```bash
    mv goot /usr/local/bin/ # Or any directory in your shell's PATH
    ```

## Usage

### Command Line Interface (CLI)

Interact with `goot` directly from your terminal for quick task operations.

* **Add a new task:**
    ```bash
    goot add "Remember to water the plants 🌿"
    ```
* **See your tasks:**
    ```bash
    goot list
    ```
* **Mark a task as done (use the ID from `goot list`):**
    ```bash
    goot done <task_id>
    ```
* **Remove a task (use the ID from `goot list`):**
    ```bash
    goot rm <task_id>
    ```
* **Launch the TUI:**
    ```bash
    goot tui
    ```

*(Explore `goot --help` for a full list of commands!)*

### Terminal User Interface (TUI)

For a more interactive task management experience, fire up the TUI:

```bash
goot # TUI mode by default
