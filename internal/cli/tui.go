package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"godesk/internal/project"
)

func newTUICommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive project workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return (&tuiSession{app: app}).run()
		},
	}
}

type tuiSession struct {
	app      *appContext
	projects []project.Project
	selected int
	message  string
	rawState string
}

func (s *tuiSession) run() error {
	if err := s.reload(); err != nil {
		return err
	}
	state, err := setRawMode()
	if err != nil {
		return fmt.Errorf("tui requires an interactive terminal: %w", err)
	}
	s.rawState = state
	defer func() {
		restoreTerminal(s.rawState)
		leaveAltScreen()
		showCursor()
	}()

	enterAltScreen()
	hideCursor()
	for {
		s.draw()
		key, err := readTUIKey()
		if err != nil {
			return err
		}
		switch key {
		case "q", "ctrl-c":
			return nil
		case "j", "down":
			s.move(1)
		case "k", "up":
			s.move(-1)
		case "r":
			if err := s.reload(); err != nil {
				s.message = err.Error()
			} else {
				s.message = "reloaded project index"
			}
		case "i":
			s.runCommand("inspect")
		case "p":
			s.runCommand("ports")
		case "h":
			s.runCommand("health")
		case "u":
			s.runCommand("up")
		case "l":
			s.runCommand("logs")
		}
	}
}

func (s *tuiSession) reload() error {
	idx, err := s.app.store.LoadIndex()
	if err != nil {
		return err
	}
	s.projects = idx.Projects
	if s.selected >= len(s.projects) {
		s.selected = len(s.projects) - 1
	}
	if s.selected < 0 {
		s.selected = 0
	}
	return nil
}

func (s *tuiSession) move(delta int) {
	if len(s.projects) == 0 {
		return
	}
	s.selected += delta
	if s.selected < 0 {
		s.selected = 0
	}
	if s.selected >= len(s.projects) {
		s.selected = len(s.projects) - 1
	}
}

func (s *tuiSession) currentProject() (project.Project, bool) {
	if len(s.projects) == 0 || s.selected < 0 || s.selected >= len(s.projects) {
		return project.Project{}, false
	}
	return s.projects[s.selected], true
}

func (s *tuiSession) draw() {
	width, height := terminalSize()
	if width < 80 {
		width = 80
	}
	if height < 20 {
		height = 20
	}
	leftWidth := 30
	rightWidth := width - leftWidth - 5

	clearScreen()
	fmt.Println("godesk tui")
	fmt.Println(strings.Repeat("=", width))
	fmt.Printf("%-*s  %s\n", leftWidth, "Projects", "Details")
	fmt.Printf("%-*s  %s\n", leftWidth, strings.Repeat("-", leftWidth), strings.Repeat("-", rightWidth))

	details := s.detailLines(rightWidth)
	rows := height - 8
	if rows < 8 {
		rows = 8
	}
	start := 0
	if s.selected >= rows {
		start = s.selected - rows + 1
	}
	for i := 0; i < rows; i++ {
		left := ""
		projectIndex := start + i
		if projectIndex < len(s.projects) {
			prefix := "  "
			if projectIndex == s.selected {
				prefix = "> "
			}
			left = prefix + s.projects[projectIndex].Name
		}
		right := ""
		if i < len(details) {
			right = details[i]
		}
		fmt.Printf("%-*s  %s\n", leftWidth, truncate(left, leftWidth), right)
	}

	fmt.Println(strings.Repeat("-", width))
	fmt.Println("[j/k or arrows] move  [r] reload  [i] inspect  [p] ports  [h] health  [u] up  [l] logs  [q] quit")
	if s.message != "" {
		fmt.Println(truncate(s.message, width))
	}
}

func (s *tuiSession) detailLines(width int) []string {
	p, ok := s.currentProject()
	if !ok {
		return []string{
			"no indexed projects",
			"run: godesk scan <root>",
			"or:  godesk roots add <root> && godesk scan",
		}
	}
	lines := []string{
		"name: " + p.Name,
		"path: " + p.Path,
		"env: " + marker(p.EnvFile),
		"compose: " + marker(p.ComposeFile),
		"lint: " + marker(p.LintCmd),
		"up: " + marker(p.UpCmd),
		"health: " + listMarker(p.HealthURLs),
		"logs: " + listMarker(p.LogFiles),
	}
	for i, line := range lines {
		lines[i] = truncate(line, width)
	}
	return lines
}

func (s *tuiSession) runCommand(name string) {
	p, ok := s.currentProject()
	if !ok {
		s.message = "no project selected"
		return
	}
	if err := restoreTerminal(s.rawState); err != nil {
		s.message = err.Error()
		return
	}
	leaveAltScreen()
	showCursor()

	args := []string{name, p.Name}
	fmt.Printf("$ godesk %s\n\n", strings.Join(args, " "))
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		cmd := exec.Command(exe, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nerror: %v\n", err)
		}
	}

	fmt.Print("\npress Enter to return to godesk tui...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	state, err := setRawMode()
	if err != nil {
		s.message = err.Error()
		return
	}
	s.rawState = state
	enterAltScreen()
	hideCursor()
	s.message = "returned from " + name
}

func readTUIKey() (string, error) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	if buf[0] == 3 {
		return "ctrl-c", nil
	}
	if buf[0] == 27 && n >= 3 && buf[1] == '[' {
		switch buf[2] {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		}
	}
	return string(buf[0]), nil
}

func setRawMode() (string, error) {
	stateCmd := exec.Command("stty", "-g")
	stateCmd.Stdin = os.Stdin
	var state bytes.Buffer
	stateCmd.Stdout = &state
	if err := stateCmd.Run(); err != nil {
		return "", err
	}

	rawCmd := exec.Command("stty", "raw", "-echo")
	rawCmd.Stdin = os.Stdin
	if err := rawCmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(state.String()), nil
}

func restoreTerminal(state string) error {
	if state == "" {
		return nil
	}
	cmd := exec.Command("stty", state)
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func terminalSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 100, 30
	}
	var rows, cols int
	if _, err := fmt.Fscanf(&out, "%d %d", &rows, &cols); err != nil {
		return 100, 30
	}
	return cols, rows
}

func enterAltScreen() {
	fmt.Print("\x1b[?1049h")
}

func leaveAltScreen() {
	fmt.Print("\x1b[?1049l")
}

func clearScreen() {
	fmt.Print("\x1b[H\x1b[2J")
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "~"
}
