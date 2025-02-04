# Benchmarks

This repository contains code to run few benchmark tests and visualize their results

## Micro Benchmarks
```
usage: micro.py [-h] --conn {agent,ssh} [--api API] [--agent AGENT] [--proxy PROXY] [--remote REMOTE] --task {bench,load} [--size SIZE] [--reps REPS] [--rate RATE]

options:
  -h, --help           show this help message and exit
  --conn {agent,ssh}   Connection Type
  --api API            [Agent] API URL
  --agent AGENT        [Agent] Agent ID
  --proxy PROXY        [SSH] Proxy Addr (user@hostname)
  --remote REMOTE      [SSH] Remote Addr (user@hostname)
  --task {bench,load}  Task to perform
  --size SIZE          Payload size (KB)
  --reps REPS          [Bench] Number of repetitions
  --rate RATE          [Load] Request rate (req/s)
```

## Examples
### Compute
```shell
./agent_amd64 -r 3.15.162.26:50051 -i agent_id_887 -s /home/ubuntu/agent_id_887.sock
```

### Router
```shell
./router -r 0.0.0.0:50051
```

### Client
#### Load
```shell
./agent_amd64 -r 3.15.162.26:50051 -i agent_id_887_load -s /home/ubuntu/agent_id_887_load.sock
./micro.py --conn=grpc --cli=/home/ubuntu/grpcsh_amd64 --sock=/home/ubuntu/agent_id_887_load.sock --peer=agent_id_887 --task=load --cmd="echo Simulating Load" --rate 1
```

#### BM
```shell

./agent_amd64 -r 3.15.162.26:50051 -i agent_id_887_bm -s /home/ubuntu/agent_id_887_bm.sock
./micro.py --conn=grpc --cli=/home/ubuntu/grpcsh_amd64 --sock=/home/ubuntu/agent_id_887_bm.sock --peer=agent_id_887 --task=cmd --cmd="echo Benchmarking Command" --reps=10 --dest=/home/ubuntu/results.jsonl
```
