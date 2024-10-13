package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"

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
					safeWriteFile(filepath.Join(gitFolder, "exclude"), []byte("# go-git ls-files --others --exclude-from=.go-git/info/exclude\n# Lines that start with '#' are comments.\n# For a project mostly in C, the following would be a good set of\n# exclude patterns (uncomment them if you want to use them):\n# *.[oa]\n# *~\n"))

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
					// &cli.BoolFlag{
					// 	Name: "stdin",
					// },
				},
				Action: func(cCtx *cli.Context) error {
					filePath := cCtx.Args().Get(0)
					write := cCtx.Bool("write")
					// stdin := cCtx.Bool("stdin")

					if filePath == "" {
						log.Fatal("file arg required")
					}

					fileBytes, err := os.ReadFile(filePath)
					if err != nil {
						log.Fatal(err)
					}

					// if stdin {
					// 	log.Fatal("TODO: support stdin")
					// }

					header := []byte("blob " + fmt.Sprint(len(fileBytes)) + "\u0000")

					fullContent := append(header, fileBytes...)

					fmt.Println(string(fullContent))

					hasher := sha1.New()
					hasher.Write(fullContent)
					hash := hex.EncodeToString(hasher.Sum(nil))

					fmt.Println(hash)

					var b bytes.Buffer
					w := zlib.NewWriter(&b)
					w.Write(fullContent)
					w.Close()
					//fmt.Print(b.String())

					//fmt.Println()
					hashRunes := []rune(hash)
					if write {
						fmt.Println("writing..")
						safeMkdir(filepath.Join(".git", "objects", string(hashRunes[:2])))
						safeWriteFile(filepath.Join(".git", "objects", string(hashRunes[:2]), string(hashRunes[2:])), b.Bytes())
					}
					fmt.Println(filepath.Join(".git", "objects", string(hashRunes[:2]), string(hashRunes[2:])))

					// r, err := zlib.NewReader(&b)
					// if err != nil {
					// 	log.Fatal(err)
					// }
					// io.Copy(os.Stdout, r)
					// r.Close()

					return nil
				},
			},
			{
				Name: "cat-file",
				Action: func(cCtx *cli.Context) error {
					filePath := cCtx.Args().Get(0)
					if filePath == "" {
						log.Fatal("file arg required")
					}

					// TODO read from .git/[0..1]/[2..40] of sha
					// TODO take 6 character input and still find correct file
					fileBytes, err := os.ReadFile(filePath)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Print(string(fileBytes))
					fmt.Println()
					var b bytes.Buffer
					b.Write(fileBytes)
					r, err := zlib.NewReader(&b)
					if err != nil {
						log.Fatal(err)
					}
					io.Copy(os.Stdout, r)
					r.Close()

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
