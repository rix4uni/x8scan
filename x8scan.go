package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rix4uni/x8scan/banner"
	"github.com/spf13/pflag"
)

// Result represents the structure for JSON output
type Result struct {
	RunningCommand string   `json:"Running_Command"`
	Method         string   `json:"method"`
	URL            string   `json:"url"`
	TransformedURL string   `json:"transformed_url"`
	Parameters     []string `json:"parameters"`
}

// X8Param represents a single parameter from x8 output
type X8Param struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Diffs       string `json:"diffs"`
	Status      int    `json:"status"`
	Size        int    `json:"size"`
	ReasonKind  string `json:"reason_kind"`
}

// X8Result represents the structure of x8 JSON output
type X8Result struct {
	Method        string    `json:"method"`
	URL           string    `json:"url"`
	Status        int       `json:"status"`
	Size          int       `json:"size"`
	FoundParams   []X8Param `json:"found_params"`
	InjectionPlace string   `json:"injection_place"`
}

func processURL(url string, method string, wordlistPath string, jsonFlag bool, verbose bool, tfilter bool, outputFile *os.File, wg *sync.WaitGroup, semaphore chan struct{}) {
	defer wg.Done()
	semaphore <- struct{}{}        // Acquire a slot
	defer func() { <-semaphore }() // Release the slot

	// Trim spaces from the method
	method = strings.TrimSpace(method)

	// Build the x8 command with JSON output format
	args := []string{"-u", url, "-X", method, "-w", wordlistPath, "--disable-progress-bar", "--reflected-only", "--output-format", "json"}
	cmd := exec.Command("x8", args...)

	// Build command string for display/logging
	command := fmt.Sprintf("x8 -u %s -X %s -w %s --disable-progress-bar --reflected-only --output-format json", url, method, wordlistPath)

	// Capture the command's output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Running command on terminal if verbose is true
	if verbose {
		fmt.Printf("Running command: %s\n", command)
	}

	// Run the command
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command for method %s: %v\n", method, err)
		fmt.Printf("Command output: %s\n", out.String())
		return
	}

	// Debug: print raw output
	if verbose {
		fmt.Printf("Raw x8 output: %s\n", out.String())
	}

	// Extract only the JSON part from the output
	// The JSON is at the end, starting with '['
	outputStr := out.String()
	lines := strings.Split(outputStr, "\n")
	var jsonLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
			jsonLine = lines[i]
			break
		}
	}

	if jsonLine == "" {
		fmt.Printf("No JSON found in x8 output for method %s\n", method)
		return
	}

	// Parse the JSON output
	var x8Results []X8Result
	if err := json.Unmarshal([]byte(jsonLine), &x8Results); err != nil {
		fmt.Printf("Error parsing JSON output for method %s: %v\n", method, err)
		fmt.Printf("JSON line was: %s\n", jsonLine)
		return
	}

	// Debug: print parsed results
	if verbose {
		fmt.Printf("Parsed %d x8 results\n", len(x8Results))
	}

	// Define the result to store the output
	var result Result
	result.RunningCommand = command
	result.Method = method
	result.URL = url

	// Process each x8 result
	for _, x8Result := range x8Results {
		if len(x8Result.FoundParams) > 0 {
			var params []string
			for _, param := range x8Result.FoundParams {
				params = append(params, param.Name)
			}

			// Construct the transformed URL with sequential x8scanN values for each parameter
			var paramStrings []string
			for i, param := range params {
				param = strings.TrimSpace(param)
				paramValue := fmt.Sprintf("x8scan%d", i+1)
				paramStrings = append(paramStrings, fmt.Sprintf("%s=%s", param, paramValue))
			}
			transformedURL := fmt.Sprintf("%s?%s", url, strings.Join(paramStrings, "&"))
			result.TransformedURL = transformedURL
			result.Parameters = params

			// Print the result as JSON if the flag is set
			if jsonFlag {
				jsonOutput, _ := json.MarshalIndent(result, "", "  ")
				writeOutput(outputFile, string(jsonOutput))
			} else if tfilter {
				// Print only the transformed URL for tool integration
				if result.TransformedURL != "" {
					writeOutput(outputFile, result.TransformedURL)
				}
			} else {
				// Print only the transformed URL
				if result.TransformedURL != "" {
					writeOutput(outputFile, fmt.Sprintf("Transformed URL [%s]: %s", method, result.TransformedURL))
				}
			}
		}
	}
}

// expandHomeDir expands ~ to the user's home directory
func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(usr.HomeDir, path[1:])
	}
	return path
}

// downloadFile downloads a file from a URL and saves it to the specified filepath
func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: received status code %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// ensureWordlistExists checks if the wordlist file exists, creates the directory if needed, and downloads the file if missing
func ensureWordlistExists(wordlistPath string) error {
	// Check if file already exists
	if _, err := os.Stat(wordlistPath); err == nil {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(wordlistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Download the wordlist file
	downloadURL := "https://raw.githubusercontent.com/rix4uni/WordList/refs/heads/main/params.txt"
	if err := downloadFile(downloadURL, wordlistPath); err != nil {
		return fmt.Errorf("failed to download wordlist: %v", err)
	}

	return nil
}

func writeOutput(outputFile *os.File, output string) {
	fmt.Println(output)
	if outputFile != nil {
		_, err := outputFile.WriteString(output + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		}
	}
}

func main() {
	// Define the flags
	methods := pflag.StringP("methods", "m", "GET", "HTTP methods to test (comma-separated)")
	wordlist := pflag.StringP("wordlist", "w", "~/.config/x8scan/params.txt", "Custom wordlist")
	concurrency := pflag.IntP("concurrency", "c", 10, "Number of concurrent URL scans")
	jsonFlag := pflag.BoolP("json", "j", false, "Output results in JSON format")
	outputFileFlag := pflag.StringP("output", "o", "", "File to save the output.")
	appendOutputFlag := pflag.StringP("append-output", "a", "", "File to append the output instead of overwriting.")
	version := pflag.Bool("version", false, "Print the version of the tool and exit.")
	silent := pflag.Bool("silent", false, "Silent mode.")
	verbose := pflag.Bool("verbose", false, "Enable verbose output for debugging purposes.")
	tfilter := pflag.BoolP("tfilter", "t", false, "Print only transformed URLs for tool integration.")
	pflag.Parse()

	if *version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	// Parse methods from the -methods flag
	methodsList := strings.Split(*methods, ",")
	if len(methodsList) == 0 {
		fmt.Println("No methods specified.")
		os.Exit(1)
	}

	// Trim whitespace from each method
	for i, method := range methodsList {
		methodsList[i] = strings.TrimSpace(method)
	}

	// Expand home directory in wordlist path and ensure wordlist exists
	wordlistPath := expandHomeDir(*wordlist)
	if err := ensureWordlistExists(wordlistPath); err != nil {
		fmt.Printf("Error setting up wordlist: %v\n", err)
		os.Exit(1)
	}

	// Open the output file for writing or appending if specified
	var outputFile *os.File
	var err error
	if *outputFileFlag != "" {
		outputFile, err = os.Create(*outputFileFlag)
	} else if *appendOutputFlag != "" {
		outputFile, err = os.OpenFile(*appendOutputFlag, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	}

	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if outputFile != nil {
			outputFile.Close()
		}
	}()

	// Create a scanner to read URLs from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Read all URLs from standard input
	var urls []string
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *concurrency)

	// Process each URL sequentially
	for _, url := range urls {
		for _, method := range methodsList {
			wg.Add(1)
			go processURL(url, method, wordlistPath, *jsonFlag, *verbose, *tfilter, outputFile, &wg, semaphore)
		}
	}

	wg.Wait()

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}
