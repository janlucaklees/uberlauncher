package todo

import (
	"bufio"
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type Skill struct {
	runtime skills.Runtime
}

func New() skills.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "todo"
}

func (s *Skill) Init(runtime skills.Runtime) {
	s.runtime = runtime

	entry := types.NewEntry(s.Name(), s.Name())
	entry.DisplayText = s.Name()
	entry.SupportsFreeText = true

	runtime.UpsertEntries([]types.Entry{entry})
}

func (s *Skill) Execute(cmd types.Command) {
	text := strings.TrimSpace(cmd.RawInput)
	if !strings.HasPrefix(text, "todo ") {
		s.runtime.ReportError(errors.New("todo input must start with 'todo '"))
		return
	}
	text = strings.TrimSpace(strings.TrimPrefix(text, "todo "))
	if text == "" {
		s.runtime.ReportError(errors.New("todo text is empty"))
		return
	}

	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		_ = loadEnvFile(filepath.Join(os.Getenv("HOME"), ".config/uberlauncher/todo.env"))
		token = os.Getenv("TODOIST_API_TOKEN")
	}
	if token == "" {
		s.runtime.ReportError(errors.New("TODOIST_API_TOKEN not set"))
		return
	}

	if err := quickAdd(token, text); err != nil {
		s.runtime.ReportError(err)
	}
}

func quickAdd(token, text string) error {
	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.todoist.com/api/v1/tasks/quick", bytes.NewReader(payload))
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
		_, _ = io.Copy(io.Discard, resp.Body)
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
