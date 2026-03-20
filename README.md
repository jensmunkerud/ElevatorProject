# ElevatorProject

Distributed elevator control system in Go, built around four core modules:

- `callhandler`
- `orderdistributor`
- `elevatorserver`
- `networking`

All other folders support these modules (hardware adapter, domain models, config, process supervision, and external libraries).

## What This Project Does

Each running instance controls one elevator and participates in a shared network worldview.  
The system:

- collects button/floor/obstruction/stop events from local hardware/simulator
- merges local and remote order/elevator state
- distributes hall calls using a cost function
- updates lights and serves assigned orders
- keeps operating under peer disconnects and process crashes (via backup process pairing)

## Running the Project

### Prerequisites

- `make` installed (used for project configuration/build)
- Go installed (project uses modules with local `replace` paths)
- Elevator simulator/hardware endpoint available on chosen port(s)

### Configure and build (recommended)

Run this from the project root:

```bash
make
```

This does the required setup/build steps:

- initializes git submodules
- builds `hall_request_assigner` for your platform
- builds the main elevator executable (`elevator` / `elevator.exe`)

### Start one instance

All flags are optional. Running without flags uses defaults from `main.go`/`config`.

```bash
./elevator
```

On Windows:

```bash
elevator.exe
```

### Available flags (all optional)

- `-port <int>`: simulator/elevator port for this instance (default: `15657`)
- `-processpair`: starts the process in backup-monitor mode instead of normal master mode
- `-masterpid <int>`: PID of the master process to monitor (used together with `-processpair`)
- `-simulator`: run in simulator mode (default: `false`)

Example using a custom port:

```bash
./elevator -port 15658
```

### Start multiple instances (example)

Run each instance in a separate terminal with unique elevator ports:

```bash
./elevator -port 15657
./elevator -port 15658
./elevator -port 15659
```

Each process uses its port as node ID in simulator mode.

### Process-pair mode (internal use)

Backup mode is started with:

```bash
./elevator -port 15657 -processpair -masterpid <PID>
```

Normally this is spawned automatically by the master instance (`processpair.SpawnAndMonitorBackup`).

## Core Architecture

### 1) `callhandler`
Local runtime state machine for one elevator.

- reads local events (buttons, floor sensor, obstruction, stop)
- drives motor/door/lamps through `controller`
- emits order/elevator-state updates to `elevatorserver`
- consumes assigned active orders from `orderdistributor`

### 2) `orderdistributor`
Hall-order assignment layer.

- receives merged worldview from `elevatorserver`
- throttles assignment updates (`config.CostFuncInterval`)
- executes `hall_request_assigner` as a subprocess
- returns this node's active order matrix back to `callhandler`

### 3) `elevatorserver`
State authority + merge logic.

- keeps latest snapshots of hall orders, cab orders, and elevator states
- merges updates from local node and network peers
- handles barrier-style order state transitions to preserve orders across failures
- publishes periodic snapshots to:
  - `callhandler` (for button lamps)
  - `orderdistributor` (for assignment)
  - `networking` (for broadcast)

### 4) `networking`
Peer discovery and worldview exchange over UDP broadcast.

- discovers peers via heartbeat (`Network-go/network/peers`)
- broadcasts local worldview (`Network-go/network/bcast`)
- decodes remote worldview and forwards it to `elevatorserver`

## Supporting Modules (Short Overview)

- `controller`: adapter to `driver-go/elevio` (motor, lamps, sensor polling)
- `elevator`, `orders`: domain models and order/state representations
- `config`: ports, timing constants, node ID generation
- `processpair`: master/backup process strategy for local crash recovery
- `libs/driver-go`: elevator I/O library
- `libs/Network-go`: UDP peer + broadcast library
- `libs/project-resources/cost_fns/hall_request_assigner`: external hall assignment executable

## Data Flow

1. `controller` emits local events
2. `callhandler` updates local behavior and pushes updates
3. `elevatorserver` merges local+network state
4. `elevatorserver` sends worldview to:
   - `networking` for broadcast
   - `orderdistributor` for assignment
   - `callhandler` for light updates
5. `orderdistributor` computes assignment and sends local active orders to `callhandler`

## Configuration

For most tuning, edit `src/config/config.go`:

- Floors/buttons: `NumFloors`, `NumButtons`
- Timing: `DoorOpenDuration`, `TravelDuration`, `ServiceTimeout`
- Network: `PeersPort`, `BroadcastPort`, `HeartbeatInterval`
- Assignment pacing: `CostFuncInterval`

Runtime behavior is also exposed as flags in `main.go` (`-port`, `-processpair`, `-masterpid`, `-simulator`).

## Testing

Run all Go tests:

```bash
go test ./...
```

## Notes

- Simulator mode is controlled by the `-simulator` flag (default is disabled).
- Cab orders are local to elevator ownership, while hall orders are distributed.
- If no peers are known yet, the system still operates with self-only worldview.
