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

					safeWriteFile(filepath.Join(gitFolder, "HEAD"), []byte("ref: refs/heads/master\n"))
					safeWriteFile(filepath.Join(gitFolder, "description"), []byte("Unnamed repository; edit this file 'description' to name the repository.\n"))

					safeMkdir(filepath.Join(gitFolder, "hooks"))
					safeMkdir(filepath.Join(gitFolder, "objects"))
					safeMkdir(filepath.Join(gitFolder, "objects", "info"))
					safeWriteFile(filepath.Join(gitFolder, "exclude"), []byte("# git ls-files --others --exclude-from=.git/info/exclude\n# Lines that start with '#' are comments.\n# For a project mostly in C, the following would be a good set of\n# exclude patterns (uncomment them if you want to use them):\n# *.[oa]\n# *~\n"))

					safeMkdir(filepath.Join(gitFolder, "objects", "pack"))
					safeMkdir(filepath.Join(gitFolder, "refs"))
					safeMkdir(filepath.Join(gitFolder, "refs", "heads"))
					safeMkdir(filepath.Join(gitFolder, "refs", "tags"))

					fmt.Println("created repo: ", cCtx.Args().Get(0))
					return nil
				},
			},
			{
				Name:    "hash-object",
				Aliases: []string{},
				Usage:   "get the SHA1 hash of an object and optionally write it to the objects database",
				Args:    true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "write",
						Aliases: []string{"w"},
					},
					&cli.BoolFlag{
						Name: "stdin",
					},
				},
				Action: func(cCtx *cli.Context) error {
					filePath := cCtx.Args().Get(0)
					write := cCtx.Bool("write")
					stdin := cCtx.Bool("stdin")

					content := ""
					if filePath != "" {
						fileBytes, err := os.ReadFile(filePath)
						if err != nil {
							log.Fatal(err)
						}
						content = string(fileBytes)
					} else if stdin {
						log.Fatal("TODO: support stdin")
					}

					header := "blob " + "\000"

					fmt.Printf("write was: %b", write)
					fmt.Printf("stdin was: %b", stdin)

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
