package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	var dstDir string
	var client http.Client // zero value of http.Client is a usable client
	flag.StringVar(&dstDir, "dst", ".", "destination directory; defaults to current directory")
	flag.DurationVar(&client.Timeout, "timeout", 1*time.Minute, "timeout for the request")
	flag.Parse()

	src := flag.Args()
	if len(src) == 0 {
		log.Fatalf("cant copy")
	}

	dstDir, err := filepath.Abs(dstDir)
	if err != nil {
		log.Fatalf("invalid destination directory: %v", err)
	}

	dst := make([]string, len(src))
	for i := range src {
		dst[i] = filepath.Join(dstDir, filepath.Base(src[i]))
	}

	errs := make([]error, len(src))

	wg := new(sync.WaitGroup)
	wg.Add(len(src))

	now := time.Now()

	for i := range src {
		i := i

		go func() {
			defer wg.Done()
			errs[i] = downloadAndSave(context.TODO(), &client, src[i], dst[i])

		}()
	}
}

func downloadAndSave(ctx context.Context, c *http.Client, url, filename string, dir string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: GET %q: %v", url, err)
	}

	resp, err := c.Do(req)

	if err != nil {
		return fmt.Errorf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status: %s", resp.Status)
	}

	dstPath := filepath.Join(*dir, filename)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, resp.Body); err != nil {
		return fmt.Errorf("copying response to file: %v", err)
	}

	return nil

}
