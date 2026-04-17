package ftp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/jlaffaye/ftp"
)

type Client struct {
	cli *ftp.ServerConn

	wdMu sync.Mutex
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of FTP client is nil")
	}

	client, err := createFtpClient(config)
	if err != nil {
		return nil, fmt.Errorf("ftp: %w", err)
	}

	return &Client{cli: client}, nil
}

func (c *Client) RawClient() *ftp.ServerConn {
	return c.cli
}

func (c *Client) ChangeDir(ctx context.Context, path string) error {
	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		c.wdMu.Lock()
		defer c.wdMu.Unlock()

		path = filepath.ToSlash(path)
		err := c.cli.ChangeDir(path)
		return struct{}{}, err
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to change directory: %w", err)
	}

	return nil
}

func (c *Client) CurrentDir(ctx context.Context) (string, error) {
	currentDir, err := wrapFuncCtx(ctx, func() (string, error) {
		c.wdMu.Lock()
		defer c.wdMu.Unlock()

		return c.cli.CurrentDir()
	})
	if err != nil {
		return "", fmt.Errorf("ftp: failed to get current directory: %w", err)
	}

	return currentDir, nil
}

func (c *Client) Delete(ctx context.Context, path string) error {
	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		path = filepath.Clean(path)
		filename := filepath.Base(path)
		if filename != path {
			c.wdMu.Lock()
			defer c.wdMu.Unlock()

			currentDir, err := c.cli.CurrentDir()
			if err != nil {
				return struct{}{}, err
			}

			targetDir := filepath.Dir(path)
			if targetDir != currentDir {
				if err := c.cli.ChangeDir(targetDir); err != nil {
					return struct{}{}, err
				}
				defer c.cli.ChangeDir(currentDir)
			}
		}

		err := c.cli.Delete(filename)
		return struct{}{}, err
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to delete file: %w", err)
	}

	return nil
}

func (c *Client) Mkdir(ctx context.Context, path string) error {
	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		c.wdMu.Lock()
		defer c.wdMu.Unlock()

		err := c.cli.MakeDir(path)
		return struct{}{}, err
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to create directory: %w", err)
	}

	return nil
}

func (c *Client) MkdirAll(ctx context.Context, path string) error {
	if path == "" || path == "." {
		return nil
	}

	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		c.wdMu.Lock()
		defer c.wdMu.Unlock()

		currentDir, err := c.cli.CurrentDir()
		if err != nil {
			return struct{}{}, err
		}

		path = filepath.ToSlash(filepath.Clean(path))
		if path == "/" || currentDir == path {
			return struct{}{}, nil
		}

		defer c.cli.ChangeDir(currentDir)

		dirs := strings.Split(path, "/")
		for i, dir := range dirs {
			if i == 0 && filepath.IsAbs(path) {
				if err := c.cli.ChangeDir("/"); err != nil {
					return struct{}{}, err
				}
			}

			if dir == "" || dir == "." {
				continue
			}

			if err := c.cli.ChangeDir(dir); err != nil {
				if err := c.cli.MakeDir(dir); err != nil {
					return struct{}{}, err
				}

				if err := c.cli.ChangeDir(dir); err != nil {
					return struct{}{}, err
				}
			}
		}

		return struct{}{}, nil
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to create directory: %w", err)
	}

	return nil
}

func (c *Client) Retrieve(ctx context.Context, path string) (*File, error) {
	file, err := wrapFuncCtx(ctx, func() (*File, error) {
		path = filepath.Clean(path)
		filename := filepath.Base(path)
		if filename != path {
			c.wdMu.Lock()
			defer c.wdMu.Unlock()

			currentDir, err := c.cli.CurrentDir()
			if err != nil {
				return nil, err
			}

			targetDir := filepath.Dir(path)
			if targetDir != currentDir {
				if err := c.cli.ChangeDir(targetDir); err != nil {
					return nil, err
				}
				defer c.cli.ChangeDir(currentDir)
			}
		}

		return c.cli.Retr(filename)
	})
	if err != nil {
		return nil, fmt.Errorf("ftp: failed to retrieve file: %w", err)
	}

	return file, err
}

func (c *Client) Store(ctx context.Context, path string, reader io.Reader, offset uint64) error {
	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		path = filepath.Clean(path)
		filename := filepath.Base(path)
		if filename != path {
			c.wdMu.Lock()
			defer c.wdMu.Unlock()

			currentDir, err := c.cli.CurrentDir()
			if err != nil {
				return struct{}{}, err
			}

			targetDir := filepath.Dir(path)
			if targetDir != currentDir {
				if err := c.cli.ChangeDir(targetDir); err != nil {
					return struct{}{}, err
				}
				defer c.cli.ChangeDir(currentDir)
			}
		}

		err := c.cli.StorFrom(filename, reader, offset)
		return struct{}{}, err
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to store file: %w", err)
	}

	return nil
}

func (c *Client) StoreString(ctx context.Context, path string, data string) error {
	reader := strings.NewReader(data)
	return c.Store(ctx, path, reader, 0)
}

func (c *Client) StoreBytes(ctx context.Context, path string, data []byte) error {
	reader := bytes.NewReader(data)
	return c.Store(ctx, path, reader, 0)
}

func (c *Client) Quit(ctx context.Context) error {
	_, err := wrapFuncCtx(ctx, func() (struct{}, error) {
		c.cli.Logout()
		return struct{}{}, c.cli.Quit()
	})
	if err != nil {
		return fmt.Errorf("ftp: failed to quit: %w", err)
	}

	return nil
}

func createFtpClient(config *Config) (*ftp.ServerConn, error) {
	client, err := ftp.Dial(resolveAddr(config.Host, config.Port))
	if err != nil {
		return nil, err
	}

	if config.Username != "" || config.Password != "" {
		if err = client.Login(config.Username, config.Password); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func resolveAddr(host string, port int) string {
	if port == 0 {
		port = defaultPort
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func wrapFuncCtx[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	type result struct {
		res T
		err error
	}

	done := make(chan result, 1)

	go func() {
		res, err := fn()
		done <- result{res, err}
	}()

	select {
	case <-ctx.Done():
		var res T
		return res, ctx.Err()
	case r := <-done:
		return r.res, r.err
	}
}
