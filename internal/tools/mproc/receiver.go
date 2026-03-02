package mproc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	xcrypto "github.com/certimate-go/certimate/pkg/utils/crypto"
)

type Receiver[TIn any, TOut any] interface {
	Receive(infile, outfile, enckey string) error
	ReceiveWithContext(ctx context.Context, infile, outfile, enckey string) error
}

type ReceiverHandler[TIn any, TOut any] func(ctx context.Context, params *TIn) (*TOut, error)

type receiver[TIn any, TOut any] struct {
	handler ReceiverHandler[TIn, TOut]
}

func (r *receiver[TIn, TOut]) Receive(infile, outfile, enckey string) error {
	return r.ReceiveWithContext(context.Background(), infile, outfile, enckey)
}

func (r *receiver[TIn, TOut]) ReceiveWithContext(ctx context.Context, infile, outfile, enckey string) error {
	if infile == "" {
		return errors.New("mproc: missing or invalid input file")
	}
	if outfile == "" {
		return errors.New("mproc: missing or invalid output file")
	}
	if enckey == "" {
		return errors.New("mproc: missing or invalid encryption key")
	}

	aesKey, err := hex.DecodeString(enckey)
	if err != nil {
		return fmt.Errorf("mproc: missing or invalid encryption key: %w", err)
	}

	aesCryptor := xcrypto.NewAESCryptor(aesKey)

	// 读取输入
	inCipherData, err := os.ReadFile(infile)
	if err != nil {
		return fmt.Errorf("mproc: failed to read input file: %w", err)
	}

	// 解密输入
	inPlainData, err := aesCryptor.CBCDecrypt(inCipherData)
	if err != nil {
		return fmt.Errorf("mproc: failed to decrypt input data: %w", err)
	}

	// 反序列化输入
	var inData TIn
	if err := json.Unmarshal(inPlainData, &inData); err != nil {
		return fmt.Errorf("mproc: failed to unmarshal input data: %w", err)
	}

	// 处理
	outData, err := r.handler(ctx, &inData)
	if err != nil {
		return err
	}

	// 序列化输出
	outPlainData, err := json.Marshal(outData)
	if err != nil {
		return fmt.Errorf("mproc: failed to marshal output data: %w", err)
	}

	// 加密输出
	outCipherData, err := aesCryptor.CBCEncrypt(outPlainData)
	if err != nil {
		return fmt.Errorf("mproc: failed to encrypt output data: %w", err)
	}

	// 写入输出
	if err := os.WriteFile(outfile, outCipherData, 0o644); err != nil {
		return fmt.Errorf("mproc: failed to write output file: %w", err)
	}

	return nil
}

// 创建并返回一个多进程指令接收器。
//
// 入参:
//   - handler: 多进程指令处理函数。
//
// 出参:
//   - 多进程指令接收器。
func NewReceiver[TIn any, TOut any](handler ReceiverHandler[TIn, TOut]) Receiver[TIn, TOut] {
	return &receiver[TIn, TOut]{handler: handler}
}
