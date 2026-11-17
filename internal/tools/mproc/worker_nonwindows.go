//go:build !windows

package mproc

func BindWorkerChildLifetime() error {
	return nil
}
