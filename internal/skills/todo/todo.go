package todo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type TodoSkill struct{}

func New() skill.Skill {
	return &TodoSkill{}
}

func (s *TodoSkill) Id() string { return "todo" }

func (s *TodoSkill) Init(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Label:      "todo",
		IsFreeText: true,
		Run: func(ec entry.Context) {
			token, ok := ctx.Config["token"].(string)
			if !ok || token == "" {
				ctx.Notifier.ReportError(fmt.Errorf("todo skill is not configured: set skills.todo.token in config.toml"))
				ec.UI.KeepOpen()
				return
			}

			payload, err := json.Marshal(map[string]string{"text": strings.TrimPrefix(ec.Input, "todo ")})
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
				return
			}

			req, err := http.NewRequest(http.MethodPost, "https://api.todoist.com/api/v1/tasks/quick", bytes.NewReader(payload))
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
				return
			}
			defer func() {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}()

			if resp.StatusCode/100 != 2 {
				ctx.Notifier.ReportError(fmt.Errorf("todoist quick add failed (%s)", resp.Status))
				ec.UI.KeepOpen()
				return
			}
			ctx.Notifier.SendNotification("Todo added")
		},
	})
}
