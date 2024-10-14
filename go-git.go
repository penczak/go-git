package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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
					gitFolder := filepath.Join(repoName, ".go-git")
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

					hashObject(filePath, write)
					return nil
				},
			},
			{
				Name: "cat-file",
				Args: true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "size",
						Aliases: []string{"s"},
					},
					&cli.BoolFlag{
						Name:    "pretty-print",
						Aliases: []string{"p"},
					},
					&cli.BoolFlag{
						Name:    "type",
						Aliases: []string{"t"},
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
					},
				},
				Action: func(cCtx *cli.Context) error {
					sha := cCtx.Args().Get(0)
					if sha == "" {
						log.Fatal("sha arg required")
					}

					// read from .git/[0..1]/[2..40] of sha
					// take 6 character SHA representation and still find correct file
					objectsSubDir := getObjectSubDirectoryPath(sha)
					files, err := os.ReadDir(objectsSubDir)
					if err != nil {
						log.Fatal("dir fatal: Not a valid object name " + sha)
					}

					filePath := ""
					for i := 0; i < len(files); i++ {
						if strings.HasPrefix(files[i].Name(), sha[2:]) {
							filePath = filepath.Join(objectsSubDir, files[i].Name())
						}
					}

					if filePath == "" {
						log.Fatal("file fatal: Not a valid object name " + sha)
					}

					fileBytes, err := os.ReadFile(filePath)
					if err != nil {
						log.Fatal(err)
					}

					var b bytes.Buffer
					b.Write(fileBytes)
					r, err := zlib.NewReader(&b)
					if err != nil {
						log.Fatal(err)
					}
					buf := new(strings.Builder)
					_, err = io.Copy(buf, r)
					if err != nil {
						log.Fatal(err)
					}
					r.Close()
					uncompressed := buf.String()
					// fmt.Println(uncompressed)
					uncompressedByNullByte := strings.Split(uncompressed, "\u0000")
					header := uncompressedByNullByte[0]
					content := uncompressedByNullByte[1]

					if cCtx.Bool("all") {
						fmt.Print(uncompressed)
					} else if cCtx.Bool("type") {
						fmt.Print(strings.Split(header, " ")[0])
					} else if cCtx.Bool("size") {
						fmt.Print(strings.Split(header, " ")[1])
					} else if cCtx.Bool("pretty-print") {
						fmt.Print(content)
					}

					return nil
				},
			},
			{
				Name:  "write-tree",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {

					// header := []byte("tree " + fmt.Sprint(len(fileBytes)) + "\u0000")

					return nil
				},
			},
			{
				Name:  "add",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {
					// pathspec := cCtx.Args().First()
					return nil
				},
			},
			{
				Name:  "status",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {

					return nil
				},
			},
			{
				Name: "update-index",
				Args: true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "add",
					},
					&cli.StringSliceFlag{
						Name: "cacheinfo",
					},
				},
				Action: func(cCtx *cli.Context) error {
					cacheinfo := cCtx.StringSlice("cacheinfo")
					fmt.Print(strings.Join(cacheinfo, "   ;   "))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func hashObject(filePath string, write bool) {
	if filePath == "" {
		log.Fatal("file arg required")
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	header := []byte("blob " + fmt.Sprint(len(fileBytes)) + "\u0000")

	fullContent := append(header, fileBytes...)

	hasher := sha1.New()
	hasher.Write(fullContent)
	sha := hex.EncodeToString(hasher.Sum(nil))

	fmt.Println(sha)

	if write {
		var b bytes.Buffer

		w := zlib.NewWriter(&b)
		w.Write(fullContent)
		w.Close()

		objectsSubDir := getObjectSubDirectoryPath(sha)
		safeMkdir(objectsSubDir)
		safeWriteFile(getObjectFilePathFromHash(sha), b.Bytes())
	}
}

func getObjectSubDirectoryPath(hash string) string {
	hashRunes := []rune(hash)
	return filepath.Join(".go-git", "objects", string(hashRunes[:2]))
}

func getObjectFilePathFromHash(hash string) string {
	dir := getObjectSubDirectoryPath(hash)
	hashRunes := []rune(hash)
	return filepath.Join(dir, string(hashRunes[2:]))
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
