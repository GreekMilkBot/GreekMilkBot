package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/GreekMilkBot/GreekMilkBot/adapter/standard"
	"github.com/GreekMilkBot/GreekMilkBot/log"
	"go.uber.org/zap"
)

func init() {
	log.SetLevel(zap.DebugLevel)
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		goMod := filepath.Join(dir, "go.mod")
		goEnv := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(goEnv); err == nil {
			fmt.Printf("Import test environment: %s\n", goEnv)
			for _, item := range strings.Split(string(data), "\n") {
				item = strings.TrimSpace(item)
				if before, after, found := strings.Cut(item, "="); found && !strings.HasPrefix(item, "#") {
					_ = os.Setenv(strings.TrimSpace(before), strings.TrimSpace(after))
				}
			}
		}
		if stat, err := os.Stat(goMod); err == nil && stat.Mode().IsRegular() {
			break
		}
		dir = filepath.Dir(dir)
	}
}
