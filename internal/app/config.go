package app

import (
	"encoding/json"
	"fmt"

	"zio-learn/internal/assets"
	"zio-learn/internal/handler"
)

func loadConfig() (handler.Config, error) {
	files := assets.Files()
	data, err := files.ReadFile("config.json")
	if err != nil {
		return handler.Config{}, fmt.Errorf("read configuration: %w", err)
	}
	var c handler.Config
	if err := json.Unmarshal(data, &c); err != nil {
		return handler.Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	c.Examples = make(map[string]string)
	for _, name := range []string{"hello.scala", "tpromise.scala"} {
		data, err := files.ReadFile("examples/" + name)
		if err != nil {
			return handler.Config{}, fmt.Errorf("read example %s: %w", name, err)
		}
		c.Examples[name] = string(data)
	}
	return c, nil
}
