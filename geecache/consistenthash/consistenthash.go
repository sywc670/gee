package consistenthash

import (
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
)

type Map struct {
	hash             hashFunc
	virtualNodeCount int
	nodes            []int
	hashMap          map[int]string
}

type hashFunc func([]byte) uint32

func NewMap(virtualNodeCount int, hash hashFunc) *Map {
	m := &Map{
		hash:             hash,
		virtualNodeCount: virtualNodeCount,
		hashMap:          make(map[int]string),
	}
	if hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < m.virtualNodeCount; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			m.nodes = append(m.nodes, hash)
			m.hashMap[hash] = node
		}
	}
	slices.Sort(m.nodes)
}

func (m *Map) Get(key string) string {
	if key == "" {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.nodes), func(i int) bool {
		return hash <= m.nodes[i]
	})
	return m.hashMap[m.nodes[index%len(m.nodes)]]
	// 取余是因为这是一个环，而Search如果找不到的话，返回值就会是len(m.nodes)，这时候应该取m.nodes[0]
}
