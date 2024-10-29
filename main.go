package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

const baseURL = "https://keiran.cc/api"

func main() {
	rootCmd := &cobra.Command{Use: "keirancli"}

	var uploadCmd = &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload a file to the site",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]
			imageUrl, err := uploadFile(filePath)
			if err != nil {
				log.Fatalf("Upload failed: %v", err)
			} else {
				fmt.Printf("File uploaded successfully. Image URL: %s\n", imageUrl)
				copyToClipboard(imageUrl)
				fmt.Println("Image URL copied to clipboard.")
			}
		},
	}

	var pasteCmd = &cobra.Command{
		Use:   "paste",
		Short: "Create a new paste",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			var title, description, content, language, expirationTime, domain string

			fmt.Print("Enter the title of the paste: ")
			title, _ = reader.ReadString('\n')
			title = strings.TrimSpace(title)

			fmt.Print("Enter the description of the paste (optional): ")
			description, _ = reader.ReadString('\n')
			description = strings.TrimSpace(description)

			fmt.Print("Enter the content of the paste: ")
			content, _ = reader.ReadString('\n')
			content = strings.TrimSpace(content)

			fmt.Print("Enter the language of the paste: ")
			language, _ = reader.ReadString('\n')
			language = strings.TrimSpace(language)

			fmt.Print("Enter the expiration time of the paste (optional): ")
			expirationTime, _ = reader.ReadString('\n')
			expirationTime = strings.TrimSpace(expirationTime)

			fmt.Print("Enter the domain of the paste (optional): ")
			domain, _ = reader.ReadString('\n')
			domain = strings.TrimSpace(domain)

			pasteUrl, err := createPaste(title, description, content, language, expirationTime, domain)
			if err != nil {
				log.Fatalf("Paste creation failed: %v", err)
				println(title, description, content, language, expirationTime, domain)
			} else {
				fmt.Printf("Paste created successfully. Paste URL: %s\n", pasteUrl)
				copyToClipboard(pasteUrl)
				fmt.Println("Paste URL copied to clipboard.")
			}
		},
	}

	var shortenCmd = &cobra.Command{
		Use:   "shorten [url]",
		Short: "Shorten a given URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			shortenedUrl, err := shortenURL(url)
			if err != nil {
				log.Fatalf("URL shortening failed: %v", err)
			} else {
				fmt.Printf("URL shortened successfully. Shortened URL: %s\n", shortenedUrl)
				copyToClipboard(shortenedUrl)
				fmt.Println("Shortened URL copied to clipboard.")
			}
		},
	}

	rootCmd.AddCommand(uploadCmd, pasteCmd, shortenCmd)
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

func uploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/upload", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	imageUrl, exists := result["imageUrl"]
	if !exists {
		return "", fmt.Errorf("imageUrl not found in response")
	}

	return imageUrl, nil
}

func createPaste(title, description, content, language, expirationTime, domain string) (string, error) {
	data := map[string]string{
		"title":          title,
		"description":    description,
		"content":        content,
		"language":       language,
		"expirationTime": expirationTime,
		"domain":         domain,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/pastes", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create paste: %s", resp.Status)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	pasteUrl, exists := result["url"]
	if !exists {
		return "", fmt.Errorf("pasteUrl not found in response")
	}

	return pasteUrl, nil
}

func shortenURL(url string) (string, error) {
	data := map[string]string{"url": url}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/shorten", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to shorten URL: %s", resp.Status)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	shortenedUrl, exists := result["shortUrl"]
	if !exists {
		return "", fmt.Errorf("shortened_url not found in response")
	}

	return shortenedUrl, nil
}

func copyToClipboard(text string) {
	err := clipboard.WriteAll(text)
	if err != nil {
		fmt.Printf("Failed to copy to clipboard: %v\n", err)
	} else {
		fmt.Println("Successfully copied to clipboard.")
	}
}
