//go:build !windows

package platform

func OwnInstance(func()) (func(), bool, error) { return func() {}, true, nil }
