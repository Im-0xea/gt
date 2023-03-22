package main

import (
	"fmt"
	"os"
	"os/signal"
	"flag"
	"math/rand"
	"time"
	"io/ioutil"
	"strings"
	"syscall"
	"unsafe"
)

func load_lang(path string) ([]string, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	fileString := string(fileBytes)
	
	words := strings.Split(fileString, "\n")
	
	return words, nil
}

func setTerminalRawMode() (*syscall.Termios, error) {
	termios := new(syscall.Termios)
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(termios)), 0, 0, 0); err != 0 {
		return nil, fmt.Errorf("failed to get terminal attributes: %v", err)
	}
	
	originalTermios := *termios
	
	termios.Iflag &^= syscall.ICRNL | syscall.INLCR | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	termios.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	termios.Cflag &^= syscall.CSIZE | syscall.PARENB
	termios.Cflag |= syscall.CS8
	
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios)), 0, 0, 0); err != 0 {
		return nil, fmt.Errorf("failed to set terminal attributes: %v", err)
	}
	
	return &originalTermios, nil
}
func resetTerminalMode(oldState *syscall.Termios) error {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0); err != 0 {
		return fmt.Errorf("failed to reset terminal attributes: %v", err)
	}
	return nil
}

func main() {
	
	var (
		help	bool
		version	bool
	)
	
	flag.BoolVar(&help, "help", false, "print usage information")
	flag.BoolVar(&version, "version", false, "print version information")
	
	flag.Parse()
	
	if help {
		flag.Usage()
		return
	}
	if version {
		fmt.Println("Version 1.0.0")
		return
	}
	
	lang, err := load_lang("en.lang")
	if err != nil {
		fmt.Printf("failed to load dictionary: %s", "en.lang")
		return
	}
	
	rand.Seed(time.Now().UnixNano())
	
	words := 10
	
	sentence := make([]int, words)
	
	for i := 0; i < words; i++ {
		sentence[i] = rand.Intn(len(lang))
	}
	for _, index := range sentence {
		fmt.Printf("%s ", lang[index])
	}
	
	fmt.Printf("\r")
	fmt.Printf("\033[6 q")
	oldState, err := setTerminalRawMode()
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return
	}
	defer resetTerminalMode(oldState)
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	
	cursor := 0
	currentWord := 0
	for {
		var b [1]byte
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			fmt.Println("Error reading stdin:", err)
			return
		}
		fmt.Printf("\033[30m")
		
		switch b[0] {
			case 0x03:
				fmt.Printf("\033[0m")
				fmt.Printf("\033[2 q")
				fmt.Println("\nReceived SIGINT. Exiting.")
				return
			case 127:
				if cursor > 0 {
					cursor -= 1
					fmt.Printf("\033[0m")
					fmt.Printf("\b%c\033[D", lang[sentence[currentWord]][cursor])
					fmt.Printf("\033[30m")
				}
				continue
		}
		if b[0] == ' ' || b[0] == '\n' {
			cursor += 1
			fmt.Printf(" ")
			continue
		}
		if cursor != -1 && b[0] == lang[sentence[currentWord]][cursor] {
			fmt.Printf("\033[32m")
		} else {
			fmt.Printf("\033[31m")
		}
		fmt.Printf("%c", rune(b[0]))
		if cursor != -1 {
			if cursor + 1 == len(lang[sentence[currentWord]]) {
				cursor = -1
				currentWord += 1
				if currentWord == words {
					fmt.Printf("\033[0m\033[2 q")
					fmt.Printf("\nfinished\n")
					return
				}
			} else {
				cursor += 1
			}
		}
	}
}
