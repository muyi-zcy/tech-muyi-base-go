package myId

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var snowFlake = SnowFlake{}

type SnowFlake struct {
	epoch     int64 // 起始时间戳
	timestamp int64 // 当前时间戳，毫秒
	centerId  int64 // 数据中心机房ID
	workerId  int64 // 机器ID
	sequence  int64 // 毫秒内序列号

	timestampBits  int64 // 时间戳占用位数
	centerIdBits   int64 // 数据中心id所占位数
	workerIdBits   int64 // 机器id所占位数
	sequenceBits   int64 // 序列所占的位数
	lastTimestamp  int64 // 上一次生成ID的时间戳
	sequenceMask   int64 // 生成序列的掩码最大值
	workerIdShift  int64 // 机器id左移偏移量
	centerIdShift  int64 // 数据中心机房id左移偏移量
	timestampShift int64 // 时间戳左移偏移量
	maxTimeStamp   int64 // 最大支持的时间

	lock sync.Mutex // 锁
}

func init() {
	snowFlake.epoch = int64(1672502400000) //设置起始时间戳
	snowFlake.centerIdBits = 5
	snowFlake.workerIdBits = 5
	snowFlake.timestampBits = 41 // 时间戳占用位数
	snowFlake.maxTimeStamp = -1 ^ (-1 << snowFlake.timestampBits)

	maxWorkerId := -1 ^ (-1 << snowFlake.workerIdBits)
	maxCenterId := -1 ^ (-1 << snowFlake.centerIdBits)

	centerId := InitDataCenterID(int64(maxCenterId))
	workerId := InitWorkerID(centerId, int64(maxWorkerId))
	snowFlake.centerId = centerId
	snowFlake.workerId = workerId

	// 参数校验
	if int(centerId) > maxCenterId || centerId < 0 {
		fmt.Printf("Center ID can't be greater than %d or less than 0", maxCenterId)
		return
	}
	if int(workerId) > maxWorkerId || workerId < 0 {
		fmt.Printf("Worker ID can't be greater than %d or less than 0", maxWorkerId)
		return
	}

	snowFlake.sequenceBits = 12 // 序列在ID中占的位数,最大为4095
	snowFlake.sequence = -1

	snowFlake.lastTimestamp = -1                                                                        // 上次生成 ID 的时间戳
	snowFlake.sequenceMask = -1 ^ (-1 << snowFlake.sequenceBits)                                        // 计算毫秒内，最大的序列号
	snowFlake.workerIdShift = snowFlake.sequenceBits                                                    // 机器ID向左移12位
	snowFlake.centerIdShift = snowFlake.sequenceBits + snowFlake.workerIdBits                           // 机房ID向左移18位
	snowFlake.timestampShift = snowFlake.sequenceBits + snowFlake.workerIdBits + snowFlake.centerIdBits // 时间截向左移22位
}

// NextId 生成下一个ID
func NextId() (int64, error) {
	snowFlake.lock.Lock() //设置锁，保证线程安全
	defer snowFlake.lock.Unlock()

	now := time.Now().UnixNano() / 1000000 // 获取当前时间戳，转毫秒
	if now < snowFlake.lastTimestamp {     // 如果当前时间小于上一次 ID 生成的时间戳，说明发生时钟回拨
		return 0, errors.New(fmt.Sprintf("Clock moved backwards. Refusing to generate myId for %d milliseconds", snowFlake.lastTimestamp-now))
	}

	t := now - snowFlake.epoch
	if t > snowFlake.maxTimeStamp {
		return 0, errors.New(fmt.Sprintf("epoch must be between 0 and %d", snowFlake.maxTimeStamp-1))
	}

	// 同一时间生成的，则序号+1
	if snowFlake.lastTimestamp == now {
		snowFlake.sequence = (snowFlake.sequence + 1) & snowFlake.sequenceMask
		// 毫秒内序列溢出：超过最大值; 阻塞到下一个毫秒，获得新的时间戳
		if snowFlake.sequence == 0 {
			for now <= snowFlake.lastTimestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		snowFlake.sequence = 0 // 时间戳改变，序列重置
	}
	// 保存本次的时间戳
	snowFlake.lastTimestamp = now

	// 根据偏移量，向左位移达到
	return (t << snowFlake.timestampShift) | (snowFlake.centerId << snowFlake.centerIdShift) | (snowFlake.workerId << snowFlake.workerIdShift) | snowFlake.sequence, nil
}

func InitDataCenterID(maxDatacenterID int64) int64 {
	id := int64(1)
	mac := getLocalHardwareAddress()
	if mac != nil {
		id = (255&(int64(mac[len(mac)-2])) | 65280&(int64(mac[len(mac)-1])<<8)) >> 6
		id %= maxDatacenterID + 1
	}
	return id
}

func InitWorkerID(datacenterID, maxWorkerID int64) int64 {
	mpid := strconv.FormatInt(datacenterID, 10) + strconv.Itoa(os.Getpid())
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(mpid))
	return int64(hash.Sum32()&0xffff) % (maxWorkerID + 1)
}

func getLocalHardwareAddress() []byte {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && !isLoopback(iface) {
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						mac := iface.HardwareAddr
						if len(mac) > 0 {
							return mac
						}
					}
				}
			}
		}
	}
	return nil
}

func isLoopback(iface net.Interface) bool {
	return iface.Flags&net.FlagLoopback != 0
}
