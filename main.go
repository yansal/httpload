package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	duration := flag.Duration("duration", time.Minute, "duration")
	goroutines := flag.Int("goroutines", 10, "goroutines")
	flag.Parse()
	url := flag.Arg(0)
	if url == "" {
		fmt.Fprintln(os.Stderr, "missing expected url")
		flag.Usage()
		os.Exit(2)
	}

	g, ctx := errgroup.WithContext(context.Background())

	sentinel := errors.New("sentinel")
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(*duration):
			return sentinel
		}
	})

	for i := 0; i < *goroutines; i++ {
		g.Go(func() error {
			return load(ctx, url)
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func load(ctx context.Context, url string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			if err := do(req); err != nil {
				return err
			}
		}
	}
}

func do(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}
	return nil
}
