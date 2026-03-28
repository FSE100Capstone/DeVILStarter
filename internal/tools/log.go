package tools

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func ProgressLogOrchestration(ctx context.Context, message string, progress int) {
	log.Println(message)
	runtime.EventsEmit(ctx, "orchestrationLog", message, progress)
}

func LogOrchestration(ctx context.Context, message string) {
	log.Println(message)
	runtime.EventsEmit(ctx, "orchestrationLog", message)
}
