package main

import (
	"DeVILStarter/internal/orchestration"
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := orchestration.Initialize(a.ctx); err != nil {
		fmt.Println("Initialization failed:", err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) CreateInfrastructure() string {
	return orchestration.CreateInfrastructure(a.ctx)
}

func (a *App) DestroyInfrastructure() {
	orchestration.DestroyInfrastructure(a.ctx)
}

func (a *App) IsInfrastructureDeployed() bool {
	return orchestration.IsInfrastructureDeployed(a.ctx)
}
