package pkg

import (
	"fmt"
	"sync"
	"time"
)

type Logger struct {
	separator   string
	lines       []string
	logMessages []string
	mu          sync.Mutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewLogger returns a pointer to a new instance of Logger, taking a
// separator that distinguishes the top lines from the regular lines
// and a variadic slice of top lines.
func NewLogger(separator string, topLines ...string) *Logger {
	logger := &Logger{
		separator:   separator,
		lines:       make([]string, len(topLines)),
		logMessages: make([]string, 0),
		stopChan:    make(chan struct{}),
	}
	copy(logger.lines, topLines)

	logger.wg.Add(1)
	go logger.updateConsole()

	return logger
}

func (l *Logger) Log(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logMessages = append([]string{message}, l.logMessages...)
	if len(l.logMessages) > 10 {
		l.logMessages = l.logMessages[:10]
	}
}

func (l *Logger) UpdateLine(lineNumber int, content string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lineNumber > 0 && lineNumber <= len(l.lines) {
		l.lines[lineNumber-1] = content
	}
}

func (l *Logger) Stop() {
	close(l.stopChan)
	l.wg.Wait()
}

func (l *Logger) updateConsole() {
	defer l.wg.Done()

	const (
		clearScreen = "\033[2J"
		moveToTop   = "\033[H"
		hideCursor  = "\033[?25l"
		showCursor  = "\033[?25h"
	)

	fmt.Print(clearScreen, hideCursor)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopChan:
			fmt.Print(showCursor)
			return
		case <-ticker.C:
			l.mu.Lock()

			fmt.Print(moveToTop)
			for _, line := range l.lines {
				fmt.Println(line)
			}

			fmt.Println(l.separator)
			for _, msg := range l.logMessages {
				fmt.Println(msg)
			}

			l.mu.Unlock()
		}
	}
}
