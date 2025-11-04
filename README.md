# go-joplin-autotagger

---

# Joplin Auto Tagger (Golang)

- This is a simple [Go](https://go.dev/) script that automatically tags your notes in [Joplin](https://joplinapp.org/)
  by analyzing the
  content of the note body and comparing it to existing tag names.
- The script uses the [Joplin Web Clipper](https://joplinapp.org/help/apps/clipper/) API and is designed to provide
  consistent tagging of large collections of notes.

## Basic capabilities

- Complete handling of collections: Uses pagination to correctly load all notes and tags, regardless of their number (more than 100).
- Smart Search: Performs a case-insensitive search, ignoring punctuation and guaranteeing matches only for whole words. This prevents mistakes like tagging "a5" inside a "b7a5c4..." hash.
- Support for complex tags:
  - Simple tag (one word): Added if the full tag name is found as a whole word.
  - Complex tag (multiple words, such as "Machine Learning"): Added if at least one word in the tag name is found as a whole word in the note text.

Security: Does not create new tags and notes, only adds existing tags to notes.

## Installation and configuration

1. Setting up the Joplin API
    - Before running the script, you must activate the [Joplin Web Clipper](https://joplinapp.org/help/apps/clipper/)
      API and obtain your token:
        - Open the [Joplin Desktop](https://joplinapp.org/help/install/#desktop-applications) application.
     - Go to Tools -> Options -> Web Clipper.
     - Enable the "Enable Web Clipper Service" option.
     - Copy the Authorization token from there.
     - Make sure Joplin Desktop is running when you run the script.
2. Go script configuration
    - Open the main.go file and replace the stub with your real token:
    ```
    // JOPLIN API CONFIGURATION
    const (
        JOPLIN_API_BASE = "http://localhost:41184"
        // !!! REPLACE THIS TOKEN WITH YOUR REAL JOPLIN API TOKEN !!!
        JOPLIN_TOKEN = "YOUR_COPIED_API_TOKEN" 
    )
    ```
3. Launch
   - Make sure you have Go (Golang) installed.
   - Save the file as main.go.
   - Open a terminal in the directory with the file.
   - Run the script: \
     ```go run main.go```

## Login for debugging (Debug Logging)

- The script contains detailed log.Printf messages that show the progress:
- Start of processing each note and its processed body (fragment).
- Tags that are already attached (and skipped).
- The exact word that was found in the text that caused the tag to be attached.
- Tags that were not found (are ignored).
- Use these messages to confirm that the tagging logic is working correctly for specific notes and tags.

TODO:
 - unit testing