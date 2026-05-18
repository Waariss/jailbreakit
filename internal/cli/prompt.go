package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type prompt struct {
	reader *bufio.Reader
	writer io.Writer
}

func newPrompt(stdin io.Reader, stdout io.Writer) *prompt {
	return &prompt{
		reader: bufio.NewReader(stdin),
		writer: stdout,
	}
}

func (p *prompt) ask(label string) string {
	for {
		fmt.Fprintf(p.writer, "%s: ", label)
		value, _ := p.reader.ReadString('\n')
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
}

func (p *prompt) askDefault(label, fallback string) string {
	fmt.Fprintf(p.writer, "%s [%s]: ", label, fallback)
	value, _ := p.reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func (p *prompt) askSecret(label string) string {
	for {
		fmt.Fprintf(p.writer, "%s: ", label)
		disableEcho := exec.Command("stty", "-echo")
		disableEcho.Stdin = os.Stdin
		_ = disableEcho.Run()
		value, _ := p.reader.ReadString('\n')
		enableEcho := exec.Command("stty", "echo")
		enableEcho.Stdin = os.Stdin
		_ = enableEcho.Run()
		fmt.Fprintln(p.writer)
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
		if _, ok := p.writer.(*os.File); ok {
			continue
		}
	}
}

func (p *prompt) confirm(label string, fallback bool) bool {
	suffix := "y/N"
	if fallback {
		suffix = "Y/n"
	}

	for {
		fmt.Fprintf(p.writer, "%s [%s]: ", label, suffix)
		value, _ := p.reader.ReadString('\n')
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return fallback
		}
		switch value {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

func (p *prompt) choose(label string, max int) int {
	for {
		fmt.Fprintf(p.writer, "%s [1-%d]: ", label, max)
		value, _ := p.reader.ReadString('\n')
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil && n >= 1 && n <= max {
			return n
		}
	}
}
