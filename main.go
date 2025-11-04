package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "time"
)

// TODO:[1] Є конфлікт з плагіном go-joplin-autotagger-by-notebook-title. тут потрібно впровадити ігнор тегів з суфіксом notebook.
// TODO:[1] unit testing.
// TODO:[1]: run on laptop start after Joplin Desktop. https://gemini.google.com/app/fc2ff7e67b8324e

// JOPLIN API CONFIGURATION
const (
    JOPLIN_API_BASE = "http://localhost:41184"
    JOPLIN_TOKEN    = "99bcbacc078b509e6a609fe7e9340d44cd115cbd5467129d2fea69380c4547f1389da9d0d7d66b06716398452379187151d6a99b0f97329de34c0bbddbb2e014"
)

// PREFIX FOR ALL TAGS CREATED BASED ON NOTEBOOK NAMES
const TAG_PREFIX = "notebook." // For ignoring

// Structures for parsing API responses
type Tag struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}

type Note struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body"`
}

type CollectionResponse struct {
    Items []json.RawMessage `json:"items"`
}

// fetchData makes a GET request to the Joplin API
// Now correctly adds an API token, regardless of the presence of other request parameters (page).
func fetchData(endpoint string) ([]byte, error) {
    // Determine whether to use '?' or '&' to add a token
    separator := "?"
    if strings.Contains(endpoint, "?") {
        separator = "&"
    }

    url := fmt.Sprintf("%s%s%stoken=%s", JOPLIN_API_BASE, endpoint, separator, JOPLIN_TOKEN)

    // Set a timeout to prevent an infinite wait
    client := http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("error executing request to %s: %w. Check that Joplin and the API are working", endpoint, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: status %d for %s", resp.StatusCode, endpoint)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }
    return body, nil
}

// getAllNotes gets a list of all notes using pagination
func getAllNotes() ([]Note, error) {
    fmt.Println("-> Get all notes (with pagination)...")
    var notes []Note
    page := 1

    for {
        // Add the page parameter to the request
        endpoint := fmt.Sprintf("/notes?page=%d&fields=id,title,body", page) // Explicitly asking for body
        body, err := fetchData(endpoint)
        if err != nil {
            return nil, err
        }

        var response CollectionResponse
        if err := json.Unmarshal(body, &response); err != nil {
            return nil, fmt.Errorf("error parsing the notes on the page %d: %w", page, err)
        }

        if len(response.Items) == 0 {
            break // Last page reached
        }

        for _, rawItem := range response.Items {
            var note Note
            // Deserialize only the required fields
            if err := json.Unmarshal(rawItem, &note); err != nil {
                log.Printf("Error parsing the note: %v", err)
                continue
            }
            notes = append(notes, note)
        }

        page++
    }

    fmt.Printf("   Found %d notes.\n", len(notes))
    return notes, nil
}

// getAllTags gets a list of all tags using pagination
func getAllTags() ([]Tag, error) {
    fmt.Println("-> Getting all tags (with pagination)...")
    var tags []Tag
    page := 1

    for {
        // Add the page parameter to the request
        endpoint := fmt.Sprintf("/tags?page=%d", page)
        body, err := fetchData(endpoint)
        if err != nil {
            return nil, err
        }

        var response CollectionResponse
        if err := json.Unmarshal(body, &response); err != nil {
            return nil, fmt.Errorf("tag parsing error on page %d: %w", page, err)
        }

        if len(response.Items) == 0 {
            break // Last page reached
        }

        for _, rawItem := range response.Items {
            var tag Tag
            if err := json.Unmarshal(rawItem, &tag); err != nil {
                log.Printf("Error parsing tag: %v", err)
                continue
            }
            tags = append(tags, tag)
        }

        page++
    }

    fmt.Printf("   Found %d tags.\n", len(tags))
    return tags, nil
}

// getNoteTags gets the IDs of the tags already attached to the note
func getNoteTags(noteID string) (map[string]bool, error) {
    endpoint := fmt.Sprintf("/notes/%s/tags", noteID)
    // Note tags do not require pagination as they rarely exceed 100
    body, err := fetchData(endpoint)
    if err != nil {
        return nil, err
    }

    var response CollectionResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("error parsing note tags: %w", err)
    }

    existingTags := make(map[string]bool)
    for _, rawItem := range response.Items {
        var tag Tag
        if err := json.Unmarshal(rawItem, &tag); err != nil {
            log.Printf("Error parsing an existing tag: %v", err)
            continue
        }
        existingTags[tag.ID] = true
    }
    return existingTags, nil
}

// associateTag attaches the tag to the note
func associateTag(noteID, tagID string) error {
    url := fmt.Sprintf("%s/tags/%s/notes?token=%s", JOPLIN_API_BASE, tagID, JOPLIN_TOKEN)

    payload := map[string]string{"id": noteID}
    jsonPayload, _ := json.Marshal(payload)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("error creating POST request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("POST request execution error: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        // A 409 Conflict status means that the tag is already attached. This is not an error for us.
        if resp.StatusCode == http.StatusConflict {
            return nil // Already attached
        }
        return fmt.Errorf("error attaching tag: status %d", resp.StatusCode)
    }

    return nil
}

func main() {
    if JOPLIN_TOKEN == "YOUR_JOPLIN_API_TOKEN" {
        log.Fatal("ERROR: Please replace stub 'YOUR_JOPLIN_API_TOKEN' with your actual API token.")
    }

    // 1. Get all tags
    allTags, err := getAllTags()
    if err != nil {
        log.Fatalf("Critical error: %v", err)
    }

    // 2. Getting all notes
    allNotes, err := getAllNotes()
    if err != nil {
        log.Fatalf("Critical error: %v", err)
    }

    // 3. Basic logic of sorting and attaching tags
    fmt.Println("\n-> Start automatically tagging notes...")
    totalTagsAdded := 0

    for i, note := range allNotes {
        // DEBUG LOG: Starting note processing
        log.Printf("--- Note processing %d/%d: '%s' (ID: %s) ---", i+1, len(allNotes), note.Title, note.ID)
        // log.Printf("--- note %v", note)
        // Get the current tags of the note to check if the tag is already attached
        existingTagIDs, err := getNoteTags(note.ID)
        if err != nil {
            log.Printf("[Skip note %d/%d '%s']: Failed to get existing tags: %v", i+1, len(allNotes), note.Title, err)
            continue
        }
        // log.Printf("   [DEBUG] note.Body %s", note.Body)
        // Preparation of the body of the note for search (conversion to lower case)
        noteBodyLower := strings.ToLower(note.Body)

        // Additional processing: Create a "cleaned" note body for WHOLE WORD searches.
        // This fixes an issue with punctuation (periods, commas, etc.) that interferes with searching.
        // 1. Replace line breaks and common punctuation marks with spaces.
        processedNoteBody := strings.ReplaceAll(noteBodyLower, "\n", " ")
        for _, punc := range []string{",", ".", ";", ":", "!", "?", "(", ")", "[", "]", "{", "}", "-", "/", "_", "`", "~", "*", "\\"} {
            processedNoteBody = strings.ReplaceAll(processedNoteBody, punc, " ")
        }
        // // 2. Normalize the spaces (reduce the multiple to one) and add spaces around the edges for a reliable search of the whole word: "word"
        processedNoteBody = " " + strings.Join(strings.Fields(processedNoteBody), " ") + " "
        // log.Printf("   [DEBUG] processedNoteBody %s", processedNoteBody)

        tagsAddedToNote := 0

        for _, tag := range allTags {
            if strings.HasPrefix(tag.Title, TAG_PREFIX) {
                // Ignoring tags based on notebook title provided by go-joplin-autotagger-by-notebook-title script.
                continue
            }

            // Check 1: Is a tag already attached to this note?
            if existingTagIDs[tag.ID] {
                // log.Printf(" [DEBUG] Tag '%s' (ID: %s) is already attached, skip.", tag.Title, tag.ID)
                continue
            }

            tagTitleLower := strings.ToLower(tag.Title)
            shouldApplyTag := false
            tagWords := strings.Fields(tagTitleLower) // Break the tag into words (simple or complex)

            // DEBUG LOG: Starting tag validation
            // log.Printf(" [DEBUG] Checking tag: '%s'. Words in tag: %d. (ID: %s)", tag.Title, len(tagWords), tag.ID)
            if len(tagWords) > 1 {
                // log.Printf(" [DEBUG] Lots of words ")
                // Complex tag: Look for AT LEAST ONE word from the tag as a WHOLE WORD in the text of the note.
                for _, word := range tagWords {
                    // log.Printf(" [DEBUG] Word %s", word)
                    // Create a search template that guarantees word limits: "word"
                    searchWord1 := " " + word + " "

                    if len(searchWord1) > 2 && strings.Contains(processedNoteBody, searchWord1) {
                        shouldApplyTag = true
                        // log.Printf("   [DEBUG] FOUND: Word '%s' from compound tag '%s' found in text.", word, tag.Title)
                        break // Found a word match, you can attach a tag
                    }
                }
            } else {
                // log.Printf("   [DEBUG] One word ")
                // log.Printf("   [DEBUG] Word %s", tagTitleLower)
                // Simple tag (one word): Look for the full tag name as a WHOLE WORD in the text of the note.
                // Create a search template that guarantees word boundaries: "tag"
                searchTagTitle1 := " " + tagTitleLower + " "

                if len(searchTagTitle1) > 2 && strings.Contains(processedNoteBody, searchTagTitle1) {
                    shouldApplyTag = true
                    // log.Printf("   [DEBUG] FOUND: Simple tag '%s' found in text.", tag.Title)
                }
            }

            // log.Printf("   [DEBUG] shouldApplyTag '%v'", shouldApplyTag)
            if shouldApplyTag {

                // Attaching the tag
                err := associateTag(note.ID, tag.ID)
                if err != nil {
                    log.Printf("[ERROR] Failed to attach tag '%s' to note'%s': %v", tag.Title, note.Title, err)
                } else {
                    fmt.Printf("   [SUCCESS] Note title '%s'.\n", note.Title)
                    fmt.Printf("   [SUCCESS] Note body '%s'.\n", note.Body)
                    fmt.Printf("   [SUCCESS] Note tagged by '%s'.\n", tag.Title)
                    tagsAddedToNote++
                    totalTagsAdded++
                }
            } else {
                // DEBUG LOG: Tag not found
                // log.Printf("   [DEBUG] Tag '%s' not found in note (no word matches).", tag.Title)
            }
        }

        if tagsAddedToNote == 0 {
            // Add logging to confirm that the note is fully processed without changes
            log.Printf("--- Note '%s': Processing completed. No new tags have been added. ---\n", note.Title)
        }
    }

    fmt.Printf("\nDONE! Added a total of %d new tags to notes.\n", totalTagsAdded)
}
