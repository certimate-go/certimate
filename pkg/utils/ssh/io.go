package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/pkg/sftp"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"

	xfilepath "github.com/certimate-go/certimate/pkg/utils/filepath"
)

// 将数据写入指定远程路径的文件。
// 如果目录不存在，将会递归创建目录。
// 如果文件不存在，将会创建该文件；如果文件已存在，将会覆盖原有内容。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - path: 文件远程路径。
//   - data: 文件数据字节数组。
//   - useSCP: 是否使用 SCP 进行传输，否则使用 SFTP。
//
// 出参:
//   - 错误。
func WriteRemote(sshCli *ssh.Client, path string, data []byte, useSCP bool) error {
	if useSCP {
		return writeRemoteWithSCP(sshCli, path, data)
	}

	return writeRemoteWithSFTP(sshCli, path, data)
}

// 与 [WriteRemote] 类似，但写入的是字符串内容。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - path: 文件远程路径。
//   - data: 文件数据字节数组。
//   - useSCP: 是否使用 SCP 进行传输，否则使用 SFTP。
//
// 出参:
//   - 错误。
func WriteRemoteString(sshCli *ssh.Client, path string, content string, useSCP bool) error {
	if useSCP {
		return writeRemoteStringWithSCP(sshCli, path, content)
	}

	return writeRemoteStringWithSFTP(sshCli, path, content)
}

// 删除指定远程路径的文件。
//
// 入参:
//   - sshCli: SSH 客户端。
//   - path: 文件远程路径。
//   - useSCP: 是否使用 SCP 进行传输，否则使用 SFTP。
//
// 出参:
//   - 错误。
func RemoveRemote(sshCli *ssh.Client, path string, useSCP bool) error {
	if useSCP {
		return errors.ErrUnsupported
	}

	return removeRemoteWithSFTP(sshCli, path)
}

func writeRemoteStringWithSCP(sshCli *ssh.Client, path string, content string) error {
	return writeRemoteWithSCP(sshCli, path, []byte(content))
}

func writeRemoteStringWithSFTP(sshCli *ssh.Client, path string, content string) error {
	return writeRemoteWithSFTP(sshCli, path, []byte(content))
}

func writeRemoteWithSCP(sshCli *ssh.Client, path string, data []byte) error {
	scpCli, err := scp.NewClientFromExistingSSH(sshCli, &scp.ClientOption{})
	if err != nil {
		return fmt.Errorf("failed to create scp client: %w", err)
	}

	reader := bytes.NewReader(data)
	err = scpCli.CopyToRemote(reader, path, &scp.FileTransferOption{})
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	return nil
}

func writeRemoteWithSFTP(sshCli *ssh.Client, path string, data []byte) error {
	sftpCli, err := sftp.NewClient(sshCli)
	if err != nil {
		return fmt.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpCli.Close()

	if err := sftpCli.MkdirAll(xfilepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	file, err := sftpCli.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	return nil
}

func removeRemoteWithSFTP(sshCli *ssh.Client, path string) error {
	sftpCli, err := sftp.NewClient(sshCli)
	if err != nil {
		return fmt.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpCli.Close()

	if err := sftpCli.MkdirAll(xfilepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	if err := sftpCli.Remove(path); err != nil {
		return fmt.Errorf("failed to remove remote file: %w", err)
	}

	return nil
}
