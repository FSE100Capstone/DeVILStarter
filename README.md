# DeVILStarter

## Overview

DeVILStarter is a Wails desktop app that provisions and tears down DeVILSona's infrastructure through a Go backend and a React + Vite frontend.

## Features

- One-click create/destroy flow with status and progress updates
- Live orchestration logs in the UI
- Desktop app built with Wails and React

## Tech Stack

- Backend: Go (Wails)
- Frontend: React + TypeScript + Vite
- UI: MUI + Emotion

## Requirements

- Go 1.24+
- Node.js 18+ and npm
- Wails CLI

## Getting Started

1) Install frontend dependencies:

```bash
cd frontend
npm install
```

2) Run in dev mode:

```bash
wails dev
```

This starts the Vite dev server with hot reload and launches the Wails app. A browser dev server for calling Go methods is available at http://localhost:34115.

## Build

Create a production build:

```bash
wails build
```

## Configuration

- Project settings: [wails.json](wails.json)
- Frontend scripts and dependencies: [frontend/package.json](frontend/package.json)

## Project Structure

- Go entry points: [main.go](main.go), [app.go](app.go)
- Backend modules: [internal](internal)
- Frontend app: [frontend/src](frontend/src)
- Wails frontend bindings: [frontend/wailsjs](frontend/wailsjs)
- Build artifacts: [build](build)

## Troubleshooting

- If the UI does not update during provisioning, ensure the Wails dev server is running and check the orchestration logs in the app.
- If Vite fails to start, delete [frontend/package.json.md5](frontend/package.json.md5) and rerun `npm install`.
