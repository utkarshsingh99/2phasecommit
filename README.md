## ShardPaxosBank - A Fault-Tolerant Distributed Transaction System

This project implements a fault-tolerant distributed transaction processing system designed to support a simple banking application.  

The system incorporates key distributed systems concepts such as **sharding, replication, Paxos consensus algorithm, and the two-phase commit protocol** to achieve scalability, fault tolerance, and correctness in transaction handling.

### Features
#### Data Sharding and Replication:
Data is partitioned across multiple shards, each replicated across servers within a cluster.
Fault tolerance is achieved by ensuring data availability even if one server in a cluster fails.

<img width="922" alt="Screenshot 2025-01-19 at 10 50 25 PM" src="https://github.com/user-attachments/assets/1ea73657-ad37-4ae8-a378-97b6e457d86a" />
<img width="922" alt="Screenshot 2025-01-19 at 10 51 39 PM" src="https://github.com/user-attachments/assets/6ac177e2-c496-4fed-9e21-43f7d2cc3f9b" />

#### Transactional Support:
*Intra-Shard Transactions*: Access data within a single shard and are processed using a modified Paxos protocol.  
*Cross-Shard Transactions*: Access data across multiple shards and are handled using the two-phase commit (2PC) protocol with Paxos for intra-shard consensus.
Consensus Mechanisms:

*Modified Paxos Protocol*: Ensures fault-tolerant ordering and execution of transactions within a shard.  
*Two-Phase Commit Protocol*: Coordinates cross-shard transactions to maintain atomicity and consistency across clusters.
<img width="922" alt="Screenshot 2025-01-19 at 10 51 05 PM" src="https://github.com/user-attachments/assets/975481c9-1e2c-4fee-9f0d-ba22c8c42e35" />

#### Fault Tolerance:
Handles fail-stop failure models with robust locking mechanisms.
Transactions abort in scenarios like insufficient balances, lock contention, or lack of quorum during consensus.

#### Performance Metrics:
Measures throughput (transactions per second) and latency (time to process a transaction).

### Outputs:
Provides the balance of any data item across all servers.
Logs committed transactions for debugging and auditing.
Performance metrics of each node in the system.

### Key Distributed Systems Concepts
1. Consensus Protocols: Paxos and Multi-Paxos for intra-shard consistency.
2. Atomicity and Consistency: Achieved using two-phase commit for cross-shard transactions.
3. Fault Tolerance: Replication within clusters and recovery mechanisms like write-ahead logs (WAL).
4. Concurrency Control: Two-phase locking to manage concurrent transactions and prevent conflicts.
5. Scalability: Partitioning and replication ensure the system can handle large-scale datasets and distributed workloads.

### Skills I developed while working on this project
1. Distributed systems design: sharding, replication, consensus protocols.  
2. Implementation of fault-tolerant protocols like Paxos and 2PC.  
3. Concurrency handling with locks and conflict resolution.  
4. RPC Communications using net/rpc package.  

The full spec doc of this project can be found [here](https://piazza.com/class_profile/get_resource/m02qzstrzw256g/m3dammzi9jp3vp).
