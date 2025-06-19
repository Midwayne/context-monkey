# Context Monkey

![GitHub release](https://img.shields.io/github/v/release/midwayne/context-monkey?style=for-the-badge)

`context-monkey` (or `como` ) is a command-line tool designed to help developers package their project's source code and structure into a single, clean text file. This is useful for providing context to LLMs.

## Features

- **All-in-One Context:** The `all` command consolidates your entire project structure and file contents into one output.
- **Selective Concatenation:** The `files` command allows you to specify individual files or use glob patterns to grab exactly what you need.
- **Project Tree View:** The `tree` command generates a clean, tree-like representation of your project's directory structure.
- **Git Aware:** Automatically uses `.gitignore` and `git ls-files` (in Git repos) to exclude unnecessary files.
- **Custom Ignores:** Provides an `--ignore` flag to specify additional files or directories to exclude.
- **Cross-Platform:** Builds and runs on Windows, macOS, and Linux.

## Installation

You can install `context-monkey` either by downloading a pre-compiled binary from the official releases (recommended) or by building it from the source.

---

### From Releases (Recommended)

This is the easiest way to get started. The project is set up to automatically build and release binaries for Windows, macOS, and Linux whenever a new version is tagged.

1.  **Download the Binary:**
    Go to the **GitHub Releases** page for this project. Find the latest release and download the appropriate binary for your operating system (`como-linux-amd64` , `como-mac-amd64` or `como-windows-amd64.exe` ).

2.  **Add to System PATH:**
    To run the `como` command from any directory, you need to place the executable in a directory that is part of your system's `PATH`.

    **For macOS & Linux:**

    ```bash
    # Rename the downloaded file to 'como' for convenience
    mv ./como-linux-amd64 /usr/local/bin/como
    # or for macOS
    mv ./como-mac-amd64 /usr/local/bin/como

    # Make the binary executable
    chmod +x /usr/local/bin/como

    # Verify the installation
    como --version
    ```

    > **Note:** `/usr/local/bin` is a standard location for user-installed executables. You can choose any other directory in your `$PATH`.

    **For Windows:**

    1.  Create a folder where you will store command-line tools, for example, `C:\bin`.
    2.  Place the downloaded `como-windows-amd64.exe` file inside this folder and rename it to `como.exe`.
    3.  Add this folder to your system's `Path` environment variable:

        - Press `Win + S` and search for "Edit the system environment variables".
        - Click the "Environment Variables" button. - Under "System variables", find and select the `Path` variable, then click "Edit".
        - Click "New" and add the path to your folder (e.g., `C:\bin`).
        - Click "OK" on all windows to save the changes.

    4.  Open a **new** terminal or PowerShell window and verify the installation:

        ```powershell
        como --version
        ```

---

### From Source (for Developers)

If you have Go installed, you can build the project from the source code.

1.  **Prerequisites:**

    - Go (version 1.24.2 or later)
    - Git
    - Make (usually pre-installed on Linux/macOS)

2.  **Clone and Build:**

    ```bash
    # Clone the repository
    git clone https://github.com/Midwayne/context-monkey.git

    cd context-monkey

    # Build the binary using the Makefile
    # This will create an executable named 'como' (or 'como.exe' on Windows)
    make build
    ```

3.  **Add to PATH (Locally Built):**

    After building, you'll have a `como` executable in the project directory. You can move it to a directory in your system's PATH, just like in the "From Releases" section.

    **For macOS & Linux:**

    ```bash
    # Move the locally built binary to a directory in your PATH
    sudo mv ./como /usr/local/bin/

    # Verify
    como --version
    ```

    **For Windows:**
    Move the `como.exe` file to the folder you created (e.g., `C:\bin`) and added to your `Path` environment variable.

## Usage

The main alias for `context-monkey` is `como`.

### `como --help`

Provides a comprehensive description of all the available functionalities. You can also use this command for specific functions, for example: `como tree --help`.

### `como all`

Gathers the project structure and the content of all relevant files.

```bash
# Print the entire project context to the console
como all -d /path/to/your/project

# Save the output to a file named 'context.txt'
como all -d /path/to/your/project -o context.txt

# Exclude all files in the 'dist' folder and all '.log' files
como all -i "dist/*,*.log"

```
