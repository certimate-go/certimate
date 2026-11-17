package ssh

import (
	"bytes"
	"fmt"
	"strings"

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

// 与 [RunCommand] 类似，但以 REPL 模式逐行执行脚本。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - commands: 待执行的脚本命令。
//
// 出参:
//   - stdout：标准输出。
//   - stderr：标准错误。
//   - err: 错误。
func RunCommandWithREPL(sshCli *ssh.Client, command string) (string, string, error) {
	lines := make([]string, 0)
	for _, line := range strings.Split(command, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	stdouts := make([]string, 0, len(lines))
	stderrs := make([]string, 0, len(lines))
	for _, line := range lines {
		stdout, stderr, err := RunCommand(sshCli, line)
		stdouts = append(stdouts, stdout)
		stderrs = append(stderrs, stderr)
		if err != nil {
			return strings.Join(stdouts, "\n"), strings.Join(stderrs, "\n"), err
		}
	}

	return strings.Join(stdouts, "\n"), strings.Join(stderrs, "\n"), nil
}
