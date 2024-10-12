package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "create a new repository",
				Action: func(cCtx *cli.Context) error {
					repoName := cCtx.Args().First()
					safeMkdir(repoName)
					gitFolder := filepath.Join(repoName, ".git")
					safeMkdir(gitFolder)
					safeMkdir(filepath.Join(gitFolder, "hooks"))
					safeMkdir(filepath.Join(gitFolder, "objects"))
					safeMkdir(filepath.Join(gitFolder, "refs"))
					safeMkdir(filepath.Join(gitFolder, "refs", "heads"))
					safeMkdir(filepath.Join(gitFolder, "refs", "tags"))

					safeWriteFile(filepath.Join(gitFolder, "HEAD"), []byte("ref: refs/heads/master\n"))

					fmt.Println("created repo: ", cCtx.Args().First())
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func safeWriteFile(filename string, c []byte) {
	err := os.WriteFile(filename, c, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func safeMkdir(relPath string) {
	err := os.Mkdir(relPath, 0777)
	if err != nil {
		log.Fatal(err)
	}
}
