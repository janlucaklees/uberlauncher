package cache

import (
	"os"
	"path/filepath"
	"uberlauncher/internal/meta"
	"uberlauncher/internal/skill"
)

type Cache struct {
	path string
}

func New() (*Cache, error) {
	path, err := meta.GetCacheRootPath()
	if err != nil {
		return nil, err
	}

	return &Cache{path: path}, nil
}

func (c *Cache) RootPath() string {
	return c.path
}

func (c *Cache) NameSpacePath(namespace string) string {
	return filepath.Join(c.RootPath(), namespace)
}

func (c *Cache) WriteFile(namespace, filename string, data []byte) error {
	dir := c.NameSpacePath(namespace)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filename), data, 0o644)
}

func (c *Cache) ReadFile(namespace, filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.NameSpacePath(namespace), filename))
}

type SkillCache struct {
	cache     *Cache
	namespace string
}

func (c *Cache) GetForSkill(skill skill.Skill) *SkillCache {
	return &SkillCache{cache: c, namespace: skill.Id()}
}

func (sc *SkillCache) Path() string {
	return sc.cache.NameSpacePath(sc.namespace)
}

func (sc *SkillCache) WriteFile(filename string, data []byte) error {
	return sc.cache.WriteFile(sc.namespace, filename, data)
}

func (sc *SkillCache) ReadFile(filename string) ([]byte, error) {
	return sc.cache.ReadFile(sc.namespace, filename)
}
