package bitmap_filter

import (
	"context"
	"hash"
	"hash/fnv"

	"github.com/redis/go-redis/v9"
)

type BitMapFilter struct {
	redisCli *redis.Client
	key      string               // 位图的键
	bitSize  int64                // 位图的大小
	hashers  []func() hash.Hash64 // 哈希函数列表
}

func NewBitMapFilterCnt(redisCli *redis.Client, key string, bitSize int64, hashCnt int) *BitMapFilter {
	filter := &BitMapFilter{
		redisCli: redisCli,
		key:      key,
		bitSize:  bitSize,
	}
	if hashCnt <= 0 {
		return NewBitMapFilterDefault(redisCli, key, bitSize)
	}
	// 初始化多个不同的哈希函数
	for i := 0; i < hashCnt; i++ {
		filter.hashers[i] = fnv.New64a
	}
	redisCli.SetBit(context.Background(), key, 0, 0) // 初始化位图
	// 设置位图的大小
	if bitSize > 0 {
		_, err := redisCli.SetBit(context.Background(), key, bitSize-1, 0).Result()
		if err != nil {
			panic(err) // 初始化位图大小失败
		}
	}
	return filter
}

func NewBitMapFilterDefault(redisCli *redis.Client, key string, bitSize int64) *BitMapFilter {
	filter := &BitMapFilter{
		redisCli: redisCli,
		key:      key,
		bitSize:  bitSize,
	}
	filter.hashers = make([]func() hash.Hash64, 10) // 默认使用10个哈希函数
	// 初始化多个不同的哈希函数
	for i := 0; i < 10; i++ {
		filter.hashers[i] = fnv.New64a
	}
	return filter
}

func NewBitMapFilter(redisCli *redis.Client, key string, bitSize int64, hasher ...func() hash.Hash64) *BitMapFilter {
	filter := &BitMapFilter{
		redisCli: redisCli,
		key:      key,
		bitSize:  bitSize,
		hashers:  hasher,
	}
	return filter
}

// 计算偏移量
func (filter *BitMapFilter) calOffsets(str string) []int64 {
	offsets := make([]int64, len(filter.hashers))
	b := []byte(str)
	for i, hasher := range filter.hashers {
		hash := hasher()
		hash.Write(b)
		offset := int64(hash.Sum64())
		offsets[i] = offset % filter.bitSize // 映射到位图中的位
	}
	return offsets
}

func (filter *BitMapFilter) Size() int64 {
	return filter.bitSize
}

func (filter *BitMapFilter) Add(str string) error {
	offsets := filter.calOffsets(str)

	// 管道批量执行
	pipe := filter.redisCli.Pipeline() // 不保证原子性 TxPipeline保证原子性
	for _, offset := range offsets {
		pipe.SetBit(context.Background(), filter.key, offset, 1)
	}
	// 执行管道命令
	_, err := pipe.Exec(context.Background())
	return err
}

func (filter *BitMapFilter) Exist(str string) bool {
	offsets := filter.calOffsets(str)
	pipe := filter.redisCli.Pipeline()
	for _, offset := range offsets {
		pipe.GetBit(context.Background(), filter.key, offset)
	}
	result, err := pipe.Exec(context.Background())
	if err != nil {
		return false
	}

	for _, res := range result {
		if cmd, ok := res.(*redis.IntCmd); ok {
			if val, err := cmd.Result(); err != nil || val == 0 {
				return false // 只要有一个位为0，就认为不存在
			}
		}
	}
	return true
}
