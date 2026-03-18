package appdata

import (
	"context"
	"log"
	"os"
	"path/filepath"
)

func CreateAppDataFolders(ctx context.Context) string {
	baseDir, err := os.UserCacheDir()
	if err != nil {
		log.Panic(err)
	}

	dirName := filepath.Join(baseDir, "devilstarter")
	if err := os.MkdirAll(dirName, 0755); err != nil {
		log.Panic(err)
	}

	return dirName
}
