package main

import (
	"bytes"
	"crypto/sha1"
	"sort"
	"strings"
)

// Hash bytes.
type Hash = [sha1.Size]byte

// Ring of nodes mapped to by shards.
type Ring interface {
	Add(node string)
	Remove(node string)
	Self() string
	NodeOf(shard uint16) string
}

// NewRing returns an instance of a consistent hash ring.
func NewRing(numReplicas uint8, self string) Ring {
	r := &ring{
		numReplicas: numReplicas,
		self:        self,
		nodes:       make(map[string]struct{}),
	}

	r.Add(self)
	return r
}

type ring struct {
	numReplicas uint8  // readonly
	self        string // readonly

	nodes  map[string]struct{}
	circle []Hash
	points map[Hash]string
}

func (r *ring) Add(node string) {
	r.nodes[node] = nothing
	r.update()
}

func (r *ring) Remove(node string) {
	delete(r.nodes, node)
	r.update()
}

func (r *ring) update() {
	count := uint(r.numReplicas) * uint(len(r.nodes))

	r.circle = make([]Hash, 0, count)
	r.points = make(map[Hash]string, count)

	hasher := sha1.New()

	for node := range r.nodes {
		for i := uint8(0); i < r.numReplicas; i++ {
			hasher.Reset()
			hasher.Write([]byte(node))
			hasher.Write([]byte{i})

			var hash Hash
			copy(hash[:], hasher.Sum(nil))

			r.points[hash] = node
			r.circle = append(r.circle, hash)
		}
	}

	sort.Slice(r.circle, func(i, j int) bool { return bytes.Compare(r.circle[i][:], r.circle[j][:]) < 0 })
}

func (r *ring) Self() string { return r.self }

func (r *ring) NodeOf(shard uint16) string {
	n := len(r.circle)
	hash := sha1.Sum([]byte{byte(shard), byte(shard >> 8)}) // binary.LittleEndian.PutUint16(...)
	i := sort.Search(n, func(i int) bool { return bytes.Compare(r.circle[i][:], hash[:]) > 0 })
	if i == n {
		i = 0
	}

	node := r.points[r.circle[i]]
	return node
}

func (r *ring) String() string {
	sb := strings.Builder{}
	sb.WriteString("RING\n")

	if len(r.nodes) == 0 {
		sb.WriteString("\t<EMPTY>\n")
		return sb.String()
	}

	for node := range r.nodes {
		sb.WriteRune('\t')
		sb.WriteString(node)
		sb.WriteRune('\n')
	}
	return sb.String()
}
