// +build nacl

// FIX IT: this is a stub to allow build on nacl.
//         Need to be implemented according nacl system facilities.

package dag1_log

import (
	"github.com/sirupsen/logrus"
)

// SyslogHook to do nothing on syslogless systems.
type SyslogHook struct {
}

// Creates a stub hook to be added to an instance of logger on syslogless systems.
func NewSyslogHook(network, raddr string, tag string) (*SyslogHook, error) {
	return &SyslogHook{}, nil
}

func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	return nil
}

func (hook *SyslogHook) Levels() []logrus.Level {
	return make([]logrus.Level, 0, 0)
}
