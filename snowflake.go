package tiga

import (
	"fmt"
	"sync"
	"time"
)

const (
	epoch         int64 = 1609459200000 // 设置起始时间 (2021-01-01 00:00:00 UTC)
	machineIDBits uint8 = 10            // 机器标识占用的位数
	sequenceBits  uint8 = 12            // 序列号占用的位数
	timestampBits uint8 = 41            // 时间戳占用的位数

	maxMachineID int64 = -1 ^ (-1 << machineIDBits) // 最大机器标识
	maxSequence  int64 = -1 ^ (-1 << sequenceBits)  // 最大序列号

	machineIDShift uint8 = sequenceBits                 // 机器标识的左移位数
	timestampShift uint8 = machineIDBits + sequenceBits // 时间戳的左移位数
)

type Snowflake struct {
	machineID int64
	sequence  int64
	lastTime  int64
	lock      sync.Mutex
}

func NewSnowflake(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", maxMachineID)
	}
	return &Snowflake{
		machineID: machineID,
	}, nil
}

func (s *Snowflake) GenerateID() int64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().UnixMilli() - epoch
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.lastTime {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now
	return (now << timestampShift) | (s.machineID << machineIDShift) | s.sequence
}
func (s *Snowflake) GenerateIDString() string {
	return fmt.Sprintf("%d", s.GenerateID())
}