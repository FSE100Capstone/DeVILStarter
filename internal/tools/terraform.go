package tools

import (
	"context"
	"log"

	version "github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

func DownloadTerraform(ctx context.Context, destinationPath string) string {
	i := install.NewInstaller()

	execPath, err := i.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product:    product.Terraform,
			Version:    version.Must(version.NewVersion("1.14.4")),
			InstallDir: destinationPath,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	return execPath
}
