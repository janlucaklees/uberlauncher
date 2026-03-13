package todo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type Skill struct{}

func New() skills.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "todo"
}

func (s *Skill) Init(runtime skills.Runtime) error {
	entry := types.NewEntry(s.Name(), s.Name())
	entry.DisplayText = s.Name()
	entry.SupportsFreeText = true

	runtime.UpsertEntries([]types.Entry{entry})
	return nil
}

func (s *Skill) Execute(cmd types.Command) error {
	text := strings.TrimSpace(cmd.RawInput)
	if !strings.HasPrefix(text, "todo ") {
		return errors.New("todo input must start with 'todo '")
	}
	text = strings.TrimSpace(strings.TrimPrefix(text, "todo "))
	if text == "" {
		return errors.New("todo text is empty")
	}

	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		_ = loadEnvFile(filepath.Join(os.Getenv("HOME"), ".config/uberlauncher/todo.env"))
		token = os.Getenv("TODOIST_API_TOKEN")
	}
	if token == "" {
		return errors.New("TODOIST_API_TOKEN not set")
	}

	return quickAdd(token, text)
}

func quickAdd(token, text string) error {
	payload := []byte(fmt.Sprintf("{\"text\":%q}", text))
	req, err := http.NewRequest(http.MethodPost, "https://api.todoist.com/api/v1/tasks/quick", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("todoist quick add failed (%s)", resp.Status)
	}
	return nil
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}
