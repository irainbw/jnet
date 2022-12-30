package fileutil

import (
	"os"
)

func NewPath(targetPath string) error {
	if _, err := os.Stat(targetPath); err != nil {
		if !os.IsExist(err) {
			mErr := os.MkdirAll(targetPath, os.ModePerm)
			if mErr != nil {
				return mErr
			}
		}
	}
	return nil
}
