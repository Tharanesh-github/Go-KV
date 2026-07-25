# Go-KV
Building a highly available, multi-node database in Go from scratch to understand distributed systems.
#  How to Run the Go-KV Cluster

This project simulates a distributed key-value store with leader election, data replication, and fault tolerance.

## Prerequisites

Before running the project, ensure you have the following installed:

- Go 1.18 or later
- Git

Verify your Go installation:

```bash
go version
```

---

## Step 1: Clone the Repository

```bash
git clone https://github.com/tharanesh/go-kv.git
cd go-kv
```

---

## Step 2: Start the Cluster (3 Nodes)

To simulate a distributed network, open **three separate terminal windows** and run one node in each terminal.

### Terminal 1 (Node A)

```bash
go run cmd/node/main.go -port 8080 -peers 8081,8082
```

### Terminal 2 (Node B)

```bash
go run cmd/node/main.go -port 8081 -peers 8080,8082
```

### Terminal 3 (Node C)

```bash
go run cmd/node/main.go -port 8082 -peers 8080,8081
```

Within a few seconds, the nodes will communicate with each other and automatically elect a leader.

One of the terminals will display a message similar to:

```text
👑 I AM THE NEW LEADER
```

---

## Step 3: Test Data Replication

Open a **fourth terminal** to act as the client.

First, identify which node became the leader (for example, `8080`).

### Store a Key-Value Pair

```bash
curl "http://localhost:8080/set?key=database&val=awesome"
```

### Retrieve the Data from a Follower

Now query a different node (for example, `8082`):

```bash
curl "http://localhost:8082/get?key=database"
```

If replication is working correctly, you'll receive the stored value even though the request was sent to a follower node.

---

## Step 4: Test Fault Tolerance

To verify automatic leader recovery:

1. Locate the terminal running the current leader.
2. Press **Ctrl + C** to stop that node.
3. Watch the remaining two terminals.

Within a few seconds, the surviving nodes will:

- Detect the missing leader
- Trigger a new leader election
- Elect a new leader automatically
- Continue serving requests without manual intervention

Try retrieving the previously stored data from one of the remaining nodes:

```bash
curl "http://localhost:8081/get?key=database"
```

The data should still be available, demonstrating successful replication and fault tolerance.

---

# Go-KV Directory Structure

This project follows standard Go conventions, separating the runnable application from the core, private business logic.

```text
Go-KV/
│
├── cmd/
│   └── node/
│       └── main.go                # The entry point. Parses flags, wires up the engine, and starts the node.
│
├── internal/
│   ├── raft/
│   │   └── consensus.go           # The Raft brain: handles leader election, timers, voting, and heartbeats.
│   │
│   ├── rpc/
│   │   └── server.go              # The network layer: HTTP handlers for client requests (SET/GET) and node-to-node communication.
│   │
│   └── storage/
│       └── engine.go              # The database engine: a thread-safe, in-memory key-value store using sync.RWMutex.
│
├── .gitignore                     # Prevents compiled binaries and editor files from being uploaded to GitHub.
├── go.mod                         # The Go module file defining the project path (github.com/tharanesh/go-kv).
└── README.md                      # Project documentation, description, and run instructions.
```
