package util

import (
	"github.com/sirupsen/logrus"
)

// SafeGo runs a function in a goroutine with panic recovery.
func SafeGo(log logrus.FieldLogger, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if log != nil {
					log.Errorf("panic recovered in goroutine: %v", r)
				} else {
					logrus.Errorf("panic recovered in goroutine: %v", r)
				}
			}
		}()
		fn()
	}()
}
