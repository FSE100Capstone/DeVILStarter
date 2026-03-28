package tools

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"DeVILStarter/internal/appdata"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func ProgressLogOrchestration(ctx context.Context, message string, progress int) {
	log.Println(message)
	appendOrchestrationLog(ctx, message)
	runtime.EventsEmit(ctx, "orchestrationLog", message, progress)
}

func LogOrchestration(ctx context.Context, message string) {
	log.Println(message)
	appendOrchestrationLog(ctx, message)
	runtime.EventsEmit(ctx, "orchestrationLog", message)
}

func appendOrchestrationLog(ctx context.Context, message string) {
	appDataDir := appdata.CreateAppDataFolders(ctx)
	logFilePath := filepath.Join(appDataDir, "orchestration.log")

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open orchestration log file: %v", err)
		return
	}
	defer file.Close()

	timestamp := time.Now().Format(time.RFC3339)
	if _, err := fmt.Fprintf(file, "%s %s\n", timestamp, message); err != nil {
		log.Printf("Failed to write orchestration log file: %v", err)
	}
}
