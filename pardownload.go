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
	var client http.Client
	flag.StringVar(&dstDir, "dst", ".", "destination directory; defaults to current directory")
	flag.DurationVar(&client.Timeout, "timeout", 1*time.Minute, "timeout for the request")
	flag.Parse()

	src := flag.Args()
	if len(src) == 0 {
		log.Fatalf("file name not provided or error reading")
	}

	dstDir, err := filepath.Abs(dstDir)
	if err != nil {
		log.Fatalf("invalid destination directory: %v", err)
	}

	dst := make([]string, len(src))
	for i := range src {
		dst[i] = filepath.Join(dstDir, fmt.Sprintf("%s.html", filepath.Base(src[i])))
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
	wg.Wait()

	log.Printf("download %d files in %v", len(src), time.Since(now))

	var errCount int
	for i := range errs {
		if errs[i] != nil {
			log.Printf("err: %s -> %s %v", src[i], dst[i], errs[i])
			errCount++
		} else {
			log.Printf("Ok: %s -> %s", src[i], dst[i])

		}
	}
	os.Exit(errCount)

}

func downloadAndSave(ctx context.Context, c *http.Client, url, dir string) error {
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

	// dstPath := filepath.Join(dir, filename)
	dstFile, err := os.Create(dir)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, resp.Body); err != nil {
		return fmt.Errorf("copying response to file: %v", err)
	}

	return nil

}
