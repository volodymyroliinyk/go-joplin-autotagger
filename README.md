# go-joplin-autotagger

A [Go](https://go.dev/) script to automatically tag notes in [Joplin](https://joplinapp.org/) based on their content. It
analyzes the text of each note and adds tags whose names match the words in the text.

## Main features

- **Full collection handling**: Uses pagination to load all notes and tags, regardless of their number.
- **"Smart" search**: Guarantees a match for whole words only, ignoring case and punctuation.
- **Support for complex tags**:
    - **Simple tag** (one word): added if an exact match is found in the text.
    - **Complex tag** (multiple words): added if at least one word from the tag name is found in the text.
- **Ignore Notebooks**: Ability to exclude notes from certain notebooks by specifying their names in the
  `ignored_notebooks.txt` file.
- **Security**: The script does not create new tags, but only adds existing ones to the notes.

## Requirements

- [Joplin Desktop](https://joplinapp.org/help/install/#desktop-applications) installed.
- Installed [Go](https://go.dev/).

## Installation and configuration

#### 1. Enable the Joplin API

Before running the script, you need to activate Joplin Web Clipper and get an access token.

1. Open the Joplin desktop application.
2. Go to `Tools` -> `Settings` -> `Web Clipper` (`Tools` -> `Options` -> `Web Clipper`).
3. Enable the ``Enable Web Clipper Service'' option.
4. Copy your **authorization token**.

**Important:** Joplin must be running while the script is running.

#### 2. Configure the access token

For security, it is recommended to pass the token via an environment variable rather than storing it in the source code.

**For Linux / macOS:**

```bash
export JOPLIN_TOKEN="YOUR_COPIED_TOKEN"
```

**For Windows (Command Prompt):**

```bash
set JOPLIN_TOKEN="YOUR_COPIED_TOKEN"
```

#### 3. Configure ignored notebooks (optional)

To exclude notes from specific notebooks, create an `ignored_notebooks.txt` file in the same directory as the script and
add the notebook names to it, each on a new line.

*Example `ignored_notebooks.txt`:*

```
# Notes from these notebooks will not be processed
personal
Work/Projects
```

## Launching

Make sure you are in the project directory and set the `JOPLIN_TOKEN` environment variable.

```bash
go run main.go
```

## Debugging

``log.Printf`' calls are commented out in the `main.go` code. If you need to see in detail how each note is processed,
which tags are there and why, you can uncomment these lines.