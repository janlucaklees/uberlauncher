package refresh_apps_cache

import "uberlauncher/internal/skills/apps"

func Run() error {
	return apps.RefreshCache()
}
