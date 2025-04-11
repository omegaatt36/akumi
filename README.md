# Akumi - SSH Target Manager TUI

Akumi is a simple Terminal User Interface (TUI) built with Go and the `bubbletea` library to help you quickly connect to your stored SSH targets.

## Features

*   Lists SSH targets defined in a configuration file.
*   Uses arrow keys (`↑`/`↓` or `k`/`j`) for navigation (cycles through the list).
*   Connects to the selected target using the system's `ssh` command upon pressing `Enter`.
*   Quits the application with `q` or `Ctrl+C`.
*   Placeholder for creating new targets (`c` key - not yet implemented).

## Installation

```sh
go install github.com/omegaatt36/akumi@latest
```

## Configuration

Akumi reads its configuration from a YAML file located at:

```
$XDG_CONFIG_PATH/akumi/config.yaml
```

If `$XDG_CONFIG_PATH` is not set, it defaults to `$HOME/.config/akumi/config.yaml`.

The configuration file structure is as follows:

```yaml
targets:
  - user: <your_user>
    host: <your_host_or_ip>
  - user: <another_user>
    host: <another_host_or_ip>
    port: <custom_port> # Optional, defaults to 22 if omitted
  # Add more targets as needed
```

**Example:**

```yaml
targets:
  - user: root
    host: 192.168.1.99
  - user: user
    host: 192.168.0.1
    port: 2222
  - user: admin
    host: example.com
```

Make sure the configuration directory and file exist before running the application for the first time if you want to pre-populate targets. If the file doesn't exist, Akumi will start with an empty list.

## Usage

1.  **Install Dependencies:**
    ```bash
    go get github.com/charmbracelet/bubbletea@latest
    go get gopkg.in/yaml.v3@latest
    ```
2.  **Run the application:**
    ```bash
    go run .
    ```

### Keybindings

*   **Up/k**: Move selection up
*   **Down/j**: Move selection down
*   **Enter**: Connect to the selected SSH target
*   **c**: (Placeholder) Intended for creating a new target
*   **q/Ctrl+C**: Quit the application

## Dependencies

*   [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
*   [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
*   Requires the `ssh` command to be available in your system's PATH.

## Future Work

*   Implement the "Create Target" functionality triggered by the 'c' key.
*   Implement target editing and deletion.
*   Add error handling and display within the TUI for failed connections or configuration issues.
