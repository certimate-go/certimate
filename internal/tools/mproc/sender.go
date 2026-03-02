package mproc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-cmd/cmd"

	xcrypto "github.com/certimate-go/certimate/pkg/utils/crypto"
)

type Sender[TIn any, TOut any] interface {
	Send(params *TIn) (*TOut, error)
	SendWithContext(ctx context.Context, params *TIn) (*TOut, error)
}

type sender[TIn any, TOut any] struct {
	command string

	logger *slog.Logger
}

func (s *sender[TIn, TOut]) Send(params *TIn) (*TOut, error) {
	return s.SendWithContext(context.Background(), params)
}

func (s *sender[TIn, TOut]) SendWithContext(ctx context.Context, params *TIn) (*TOut, error) {
	// 生成随机密钥
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("mproc: failed to generate aes key: %w", err)
	}

	aesCryptor := xcrypto.NewAESCryptor(aesKey)

	// 准备临时输入文件
	tempIn, err := os.CreateTemp("", "certimate.mprocin_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("mproc: failed to create temp input file: %w", err)
	} else {
		inPlainData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("mproc: failed to marshal input data: %w", err)
		}

		inCipherData, err := aesCryptor.CBCEncrypt(inPlainData)
		if err != nil {
			return nil, fmt.Errorf("mproc: failed to encrypt input data: %w", err)
		}

		if _, err := tempIn.Write(inCipherData); err != nil {
			return nil, fmt.Errorf("mproc: failed to write input file: %w", err)
		}

		tempIn.Close()
	}
	defer os.Remove(tempIn.Name())

	// 准备临时输出文件
	tempOut, err := os.CreateTemp("", "certimate.mprocout_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("mproc: failed to create temp output file: %w", err)
	} else {
		tempOut.Close()
	}
	defer os.Remove(tempOut.Name())

	// 准备临时错误文件
	tempErr, err := os.CreateTemp("", "certimate.mprocerr_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("mproc: failed to create temp error file: %w", err)
	} else {
		tempErr.Close()
	}
	defer os.Remove(tempOut.Name())

	// 初始化子进程
	done := make(chan struct{})
	mcmd := cmd.NewCmdOptions(cmd.Options{Buffered: false, Streaming: true},
		s.getEntrypoint(),
		"intercmd",
		s.command,
		"--in", tempIn.Name(),
		"--out", tempOut.Name(),
		"--err", tempErr.Name(),
		"--enckey", hex.EncodeToString(aesKey),
	)
	go func() {
		defer close(done)
		for mcmd.Stdout != nil || mcmd.Stderr != nil {
			select {
			case line, open := <-mcmd.Stdout:
				{
					if !open {
						mcmd.Stdout = nil
						continue
					}

					if s.logger != nil {
						print := s.logger.Info

						// split log level prefix for those vendor packages:
						// - github.com/go-acme/lego: INFO, WARN
						if strings.HasPrefix(line, "[INFO] ") {
							line = strings.TrimPrefix(line, "[INFO] ")
							print = s.logger.Info
						} else if strings.HasPrefix(line, "[WARN] ") {
							line = strings.TrimPrefix(line, "[WARN] ")
							print = s.logger.Warn
						}

						print(line)
					}
				}

			case line, open := <-mcmd.Stderr:
				{
					if !open {
						mcmd.Stderr = nil
						continue
					}

					if s.logger != nil {
						print := s.logger.Error

						// split log level prefix for those vendor packages:
						// - github.com/nrdcg/desec: DEBUG
						if strings.Contains(line, "[DEBUG] ") {
							line = strings.SplitN(line, "[DEBUG] ", 2)[1]
							print = s.logger.Debug
						}

						print(line)
					}
				}
			}
		}
	}()

	// 等待子进程退出
	<-mcmd.Start()
	<-done
	if err := mcmd.Status().Error; err != nil {
		return nil, fmt.Errorf("mproc: failed to exec child process: %w", err)
	}

	// 读取输出
	outCipherData, err := os.ReadFile(tempOut.Name())
	if err != nil {
		return nil, fmt.Errorf("mproc: failed to read output file: %w", err)
	} else {
		errData, _ := os.ReadFile(tempErr.Name())
		if len(errData) > 0 {
			return nil, errors.New(string(errData))
		}
	}

	// 解密输出
	outPlainData, err := aesCryptor.CBCDecrypt(outCipherData)
	if err != nil {
		return nil, fmt.Errorf("mproc: failed to decrypt output data: %w", err)
	}

	// 反序列化输出
	var outData TOut
	if err := json.Unmarshal(outPlainData, &outData); err != nil {
		return nil, fmt.Errorf("mproc: failed to unmarshal output data: %w", err)
	}

	return &outData, nil
}

func (s *sender[TIn, TOut]) getEntrypoint() string {
	executable, err := os.Executable()
	if err != nil {
		executable = os.Args[0]
	}
	return executable
}

// 创建并返回一个多进程指令发送器。
//
// 入参:
//   - command: 多进程指令命令。需要先注册为 `intercmd [command]` 命令行。
//   - logger: 日志记录器，将重定向多进程的标准输出流和标准错误流到该日志记录器中。
//
// 出参:
//   - 多进程指令发送器。
func NewSender[TIn any, TOut any](command string, logger *slog.Logger) Sender[TIn, TOut] {
	return &sender[TIn, TOut]{command: command, logger: logger}
}
