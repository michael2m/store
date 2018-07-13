# Design

**Observations working with AWS EFS (NFS)**:
- Open/Close and Lock/Unlock:
  - Are SLOW compared to Read/Write from a compute node to EFS
  - They BOTTLENECK the IOPS throughput on a compute node to EFS
- Improve by distributing IOPS among:
  - Multiple EFS file systems
  - Multiple compute nodes

**Goal**: minimize open/close and lock/unlock

**Solution**:
- Distribute files among shards, assuming #files >> #shards
- Distribute shards among compute nodes, assuming #shards >> #nodes
- Manage locking at shard granularity, rather than files -> Minimize lock/unlock
  - Acquire shard lock lazily (from a node's perspective), i.e. upon first "access" to a file in the shard
  - Release shard lock eagerly (from a node's perspective), i.e. when the shard redistributes to another node
- Cache files in compute nodes -> Minimize open/close
  - Cache file lazily (from a node's perspective), i.e. upon first "access" to a file
  - Evict file eagerly (from a node's perspective), i.e. when file's shard redistributes to another node

---

### File F maps to shard S by hashmod: SHA1(F) % #shards
A file always maps to the same shard, using hash-modulo mapping of file paths to shards.
The number of shards is fixed / constant (upfront).

### Shard S maps to node N by consistent hash ring: Consistent-SHA1(#replicas, [nodes], S)
A shard always maps to a single node, using a consistent-hash ring mapping shard numbers to nodes.

### Node N
Node membership is gossiped by the SWIMM protocol. No reliance on AWS API limits to lookup nodes.
Although membership is eventually consistent, shard locking is strongly consistent (by AWS EFS).

---

**FS API**:
+ Open(path string) (File, error) // assume create-or-open, read-write access and 0644 mode
+ Close(path) error
+ Remove(path) error

---

**Modules**:
+ Memberlist (Hashicorp)
+ Ring (consistent-hash)
+ Sharder (hashmod, locking)
+ Cache (map or LRU?)
+ FS
+ REST API (Gin?)