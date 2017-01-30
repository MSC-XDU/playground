/*
使用一个简单实现的 Bloom Filter 实现了分享ID 生成器
ID 生成的过程：
	1.随机生成 3 个 byte
	2.将随 bytes 进行 hash 产生 Bloom Filter 所需要的 hash 数目
	3.由 Bloom Filter 进行过滤，如不存在则插入过滤器，否则返回 ErrKeyExist
	4.将随机 bytes 转换成字符串 ID 返回

注意:
	因为采取了随机字符串，当 ID 分配数接近 1000W 时性能会开始剧烈恶化。
	可以通过修改代码提高元素空间数解决。

性能:
	因为预计使用量不大，所以没有做太仔细的优化。简单的测试结果如下

	单线程连续请求:
		产生约 1400W ID 99% 的请求延迟低于 1ms。
	使用 wrk2 进行 5 线程， 10 连接， 1000 QPS 测试:
		99% 请求延迟低于 100ms，99.9% 请求延时低于 380ms。
	使用 wrk2 进行 6 线程， 40 连接， 80 QPS 测试:
		99% 请求延迟低于 6ms， 99.9% 请求延时低于 9ms。

	以上测试均在 2014 款 MacBook Pro 15 上进行。
*/
package share

import (
	"crypto/sha512"
	"encoding/binary"
	"math/rand"
	"time"
	"unsafe"

	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

const (
	randBytes = 3 // 随机 bytes 数

	bloomK = 8                    // Bloom Filter hash 函数个数
	bloomN = 1 << (randBytes * 8) // Bloom Filter 总元素数
	bloomM = bloomN * 10          // Bloom Filter bitmap 长度
	// 在以上参数下 bloomP ≈ 0.00846

	// 通过将 bitmap 切分，简单处理并发性能差的问题
	bloomBitSections    = 1 << 13
	bloomBitSectionElms = bloomM / bloomBitSections

	bloomBitMapDB = 0
)

// 进制转换字母表
var digits = [62]rune{
	'0', '1', '2', '3', '4', '5', '6', '7', '8',
	'9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
	'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q',
	'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
	'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R',
	'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

var redisAddr string

// Bloom Filter 所用的 Redis connection pool
var bfPool = redis.Pool{
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", redisAddr, redis.DialDatabase(bloomBitMapDB))
	},
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
}

// Errors
var (
	ErrKeyExist     = errors.New("this key exist")
	ErrTooManyRetry = errors.New("too many retry")
	ErrBadValue     = errors.New(`id or num cannot be "" or 0`)
)

func init() {
	// 设置 rand seed
	rand.Seed(time.Now().Unix())

	redisAddr = os.Getenv("SHARE_ID_REDIS_ADDR")
	if redisAddr == "" {
		panic("no redis addr")
	}
}

// Bloom Filter hash 函数
func bloomHash(raw []byte) [bloomK]uint32 {
	// 当前选取的方案是使用一个输出范围很大的函数进行 hash
	// 然后切分 hash 结果，产生 bloomK 个 hash 值

	// 考虑当前选取的 bloomK 和 BloomM 故选取 SHA-512/256 算法
	sha := sha512.Sum512_256(raw)

	// 因为 SHA-512/256 返回32 bytes，故可以直接将其转化为8个 uint32 的数组
	result := (*[bloomK]uint32)(unsafe.Pointer(&sha))

	for i, v := range result {
		// 消除越界
		result[i] = v % bloomM
	}

	return *result
}

func mightContains(conn redis.Conn, hashes *[bloomK]uint32) (bool, error) {
	var err error

	// 为了方便插入的实现，故不在事务中查询
	for _, hash := range hashes {
		if err = conn.Send("GETBIT", hash/bloomBitSectionElms, hash%bloomBitSectionElms); err != nil {
			return true, err
		}
	}

	if err = conn.Flush(); err != nil {
		return true, err
	}

	for range hashes {
		s, err := redis.Int(conn.Receive())
		if err != nil {
			return true, err
		}
		// 元素不存在
		if s == 0 {
			return false, nil
		}
	}

	// 元素可能存在
	return true, nil
}

// Bloom Filter 查询操作
func bloomMightContains(hashes [bloomK]uint32) (bool, error) {
	conn := bfPool.Get()
	defer conn.Close()
	return mightContains(conn, &hashes)
}

// Bloom Filter 添加操作
func bloomAdd(hashes [bloomK]uint32) error {
	conn := bfPool.Get()
	defer conn.Close()
	var err error

	// 因为检查与写入之间可能有其他客户端执行插入，需要反复重试直到成功
	for {
		// 因为 bitmap 经过切分，故需要 WATCH 所有的 section
		for _, hash := range hashes {
			if err = conn.Send("WATCH", hash/bloomBitSectionElms); err != nil {
				return err
			}
		}
		conn.Do("")

		e, err := mightContains(conn, &hashes)
		if err != nil {
			return err
		}
		if e {
			return ErrKeyExist
		}

		// 执行 Redis 事务，在事务中设置所有的 bit 位
		// 在循环开始时已经 WATCH 所有有关的 bitmap section
		// 如果有任何一个其他的 ID 生成操作修改了相关 bitmap section 该事务将失败
		// 这种情况下就会重新执行循环，检查当前的 key 时候被其它生成操作添加
		if err = conn.Send("MULTI"); err != nil {
			return err
		}

		// Bloom Filter 添加操作
		for _, hash := range hashes {
			if err = conn.Send("SETBIT", hash/bloomBitSectionElms, hash%bloomBitSectionElms, 1); err != nil {
				return err
			}
		}

		// 执行事务
		if _, err = redis.Values(conn.Do("EXEC")); err != nil {
			if err == redis.ErrNil {
				// 其他数据插入，重新检查
				continue
			}
			// 其他错误，返回
			return err
		}

		// 成功插入
		return nil
	}
}

// 产生随机 ID
func NewID() (string, error) {
	var raw [randBytes]byte
	var i uint64
	for {
		rand.Read(raw[:])
		i = bytesToI(raw)
		if i == 0 {
			continue
		}
		err := bloomAdd(bloomHash(raw[:]))
		if err != nil {
			if err != ErrKeyExist {
				return "", err
			}
			continue
		}
		break
	}
	return ItoID(i)
}

// 随机 bytes 转化成 uint64。
// 当随机 bytes 长度改变时注意修改此函数实现
func bytesToI(raw [randBytes]byte) uint64 {
	var bytes [4]byte
	copy(bytes[1:], raw[:])
	return uint64(binary.BigEndian.Uint32(bytes[:]))
}

// 将一个 uint64 转化成 ID
func ItoID(u uint64) (string, error) {
	if u == 0 {
		return "", ErrBadValue
	}
	a := [11]rune{}
	var i int
	for i = 10; u != 0; i-- {
		if i < 0 {
			return "", errors.New("out of range")
		}
		q := u / 62
		a[i] = digits[u - q*62]
		u = q
	}
	return string(a[i + 1:]), nil
}

// 将 ID  转化成 uint64
func IDtoI(id string) (uint64, error) {
	runes := []rune(id)
	l := len(runes)
	if l == 0 {
		return 0, ErrBadValue
	}
	var result uint64
	for i, r := range runes {
		var v rune
		switch {
		case r >= '0' && r <= '9':
			v = r - '0'
		case r >= 'a' && r <= 'z':
			v = r - 'a' + 10
		case r >= 'A' && r <= 'Z':
			v = r - 'A' + 36
		default:
			return 0, errors.New("id contain bad char")
		}
		result += uint64(v) * pow(62, l-i-1)
	}

	return result, nil
}

func pow(x uint64, n int) uint64 {
	if n == 0 {
		return 1
	}
	var result uint64 = 1

	for n > 0 {
		if n&1 != 0 {
			result *= x
		}
		x *= x
		n >>= 1
	}

	return result
}
