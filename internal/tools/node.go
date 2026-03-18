package tools

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsurePortableNode(ctx context.Context, destinationPath string) (string, string) {
	if runtime.GOOS != "windows" {
		log.Fatal("portable Node download is implemented for Windows in this example")
	}

	// Keep the version explicit so builds are reproducible.
	nodeVersion := "v20.11.1"
	zipName := fmt.Sprintf("node-%s-win-x64.zip", nodeVersion)
	zipURL := fmt.Sprintf("https://nodejs.org/dist/%s/%s", nodeVersion, zipName)
	zipPath := filepath.Join(destinationPath, zipName)

	if err := os.MkdirAll(destinationPath, 0755); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		err = downloadFile(ctx, zipURL, zipPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := unzipSafe(zipPath, destinationPath); err != nil {
		log.Fatal(err)
	}

	// Node zip extracts to a versioned folder.
	nodeRoot := filepath.Join(destinationPath, fmt.Sprintf("node-%s-win-x64", nodeVersion))
	npmCmd := filepath.Join(nodeRoot, "npm.cmd")
	return nodeRoot, npmCmd
}

func RunNpmInstallForLambdaFolders(ctx context.Context, npmCmdPath, lambdaRoot string) error {
	entries, err := os.ReadDir(lambdaRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packageJSON := filepath.Join(lambdaRoot, entry.Name(), "package.json")
		if _, err := os.Stat(packageJSON); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		projectDir := filepath.Join(lambdaRoot, entry.Name())
		if err := runNpmInstall(ctx, npmCmdPath, projectDir); err != nil {
			return fmt.Errorf("npm install failed in %s: %w", projectDir, err)
		}

		zipPath := filepath.Join(lambdaRoot, entry.Name()+".zip")
		if err := zipDirectory(projectDir, zipPath); err != nil {
			return fmt.Errorf("zip failed for %s: %w", projectDir, err)
		}
	}

	return nil
}

func runNpmInstall(ctx context.Context, npmCmdPath, projectDir string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("npm install is configured for Windows in this example")
	}

	cmd := ExecCommandContext(ctx, npmCmdPath, "install")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func downloadFile(ctx context.Context, url, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func unzipSafe(zipPath, destination string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		targetPath := filepath.Join(destination, cleanName)
		if !strings.HasPrefix(targetPath, destination+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		in, err := file.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(targetPath)
		if err != nil {
			in.Close()
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		out.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func zipDirectory(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		fileInfo, err := d.Info()
		if err != nil {
			file.Close()
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		entry, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(entry, file)
		file.Close()
		return err
	})
}
