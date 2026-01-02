package main

import (
	"flag"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

// Load the user32.dll library lazily (only when actually used)
// This DLL contains Windows UI functions including message boxes
var (
	user32                = syscall.NewLazyDLL("user32.dll")
	// Get a reference to the MessageBoxTimeoutW function
	// This is an undocumented Windows API that adds timeout capability
	procMessageBoxTimeout = user32.NewProc("MessageBoxTimeoutW")
)

// MessageBoxTimeout wraps the Windows API call to show a message box with auto-close
// Parameters:
//   hwnd: Handle to parent window (0 for none)
//   text: The message to display
//   caption: The title bar text
//   uType: Button and icon type constants (MB_OK, MB_YESNO, etc.)
//   wLanguageID: Language identifier (0 for system default)
//   milliseconds: Time in ms before auto-close (0 for no timeout)
// Returns: Button clicked (IDYES, IDNO, IDTIMEOUT, etc.)
func MessageBoxTimeout(hwnd uintptr, text, caption string, uType uint32, wLanguageID uint16, milliseconds uint32) int {
	// Call the Windows API function
	// We need to convert Go strings to UTF-16 pointers that Windows expects
	ret, _, _ := procMessageBoxTimeout.Call(
		hwnd,
		// Convert text to UTF-16 pointer for Windows
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		// Convert caption to UTF-16 pointer for Windows
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		uintptr(uType),        // Button/icon type
		uintptr(wLanguageID),  // Language (0 = default)
		uintptr(milliseconds), // Timeout in milliseconds
	)
	// Return the button ID that was clicked (or timeout code)
	return int(ret)
}

func main() {
	// Define command-line flags for configuration
	msg := flag.String("msg", "Hello!", "Message to display")
	timeout := flag.Int("timeout", 5, "Time before auto-close (seconds)")
	exePath := flag.String("exe", "", "Path to executable to run if Yes is clicked")
	yesNo := flag.Bool("yesno", false, "Show Yes/No buttons instead of OK")
	autoYes := flag.Bool("autoyes", false, "Choose Yes automatically if timeout occurs")
	
	// Parse the command-line flags
	flag.Parse()

	fmt.Printf("Showing message box: %s\n", *msg)

	// Windows MessageBox button type constants
	const MB_OK = 0x0    // Single OK button
	const MB_YESNO = 0x04 // Yes and No buttons
	
	// Windows MessageBox return value constants
	const IDYES = 6     // User clicked Yes
	const IDTIMEOUT = 5 // Message box timed out

	// Check if we're in Yes/No mode (interactive with executable)
	if *yesNo {
		// Validate that an executable path was provided
		if *exePath == "" {
			fmt.Println("Please provide an executable to run using -exe flag when using Yes/No box.")
			return
		}

		// Show the Yes/No message box with timeout
		ret := MessageBoxTimeout(0, *msg, "Go MsgBox", MB_YESNO, 0, uint32(*timeout*1000))

		// Check if user clicked Yes OR if timeout occurred with auto-yes enabled
		if ret == IDYES || (ret == IDTIMEOUT && *autoYes) {
			fmt.Println("Yes clicked or auto-yes triggered, running executable:", *exePath)
			
			// Create a command to execute the specified program
			cmd := exec.Command(*exePath)
			
			// Start the executable (non-blocking)
			err := cmd.Start()
			if err != nil {
				fmt.Println("Failed to start executable:", err)
				return
			}
			fmt.Println("Executable started successfully.")
		} else {
			// User clicked No or timeout without auto-yes
			fmt.Println("No clicked or timeout without auto-yes.")
		}
	} else {
		// Simple OK message box mode
		// The exePath flag is ignored in this mode
		MessageBoxTimeout(0, *msg, "Go MsgBox", MB_OK, 0, uint32(*timeout*1000))
		fmt.Println("OK message box closed.")
	}
}
