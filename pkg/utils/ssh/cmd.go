package ssh

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// 执行远程脚本命令，并返回执行后标准输出和标准错误。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - command: 待执行的脚本命令。
//
// 出参:
//   - stdout：标准输出。
//   - stderr：标准错误。
//   - err: 错误。
func RunCommand(sshCli *ssh.Client, command string) (string, string, error) {
	session, err := sshCli.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	stdoutBuf := bytes.NewBuffer(nil)
	session.Stdout = stdoutBuf
	stderrBuf := bytes.NewBuffer(nil)
	session.Stderr = stderrBuf
	err = session.Run(command)
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("failed to execute ssh command: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// 逐行执行远程脚本命令，每行命令使用独立的会话通道，并返回执行后标准输出和标准错误。
// 适用于不提供 Shell、每个会话仅接受单条命令的嵌入式设备。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - commands: 待执行的脚本命令数组。
//
// 出参:
//   - stdout：标准输出。
//   - stderr：标准错误。
//   - err: 错误。
func RunCommands(sshCli *ssh.Client, commands []string) (string, string, error) {
	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)
	for _, command := range commands {
		stdout, stderr, err := RunCommand(sshCli, command)
		stdoutBuf.WriteString(stdout)
		stderrBuf.WriteString(stderr)
		if err != nil {
			return stdoutBuf.String(), stderrBuf.String(), err
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}
