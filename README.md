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

examples:
./micro.py --conn=ssh --proxy=ubuntu@3.142.234.94 --remote=ubuntu@3.142.234.94 --task=bench --size=64 --reps=100
./micro.py --conn=ssh --proxy=ubuntu@3.142.234.94 --remote=ubuntu@3.142.234.94 --task=load --size=64 --rate=1

```