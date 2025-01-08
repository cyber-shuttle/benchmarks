#!/usr/bin/env python3

# MICRO BENCHMARKS

import time
import random
import argparse
import rich
import rich.progress
from sdk import AgentSDK, SSHSDK, SDK

def test_latency(sdk: SDK, task: str, size_kb: int, reps: int, warmup=0.1) -> float:
    """
    Measure the round-trip latency
    
    """
    if task == "uload":
        fn = lambda: sdk.uload(size_kb)
    elif task == "dload":
        fn = lambda: sdk.dload(size_kb)
    latencies = []
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task("counter", total=reps)
    r = int(reps * warmup)
    counter = 0
    while counter != reps:
        start_time = time.time_ns()
        fn()
        end_time = time.time_ns()
        progress.update(prog_bar, advance=1)
        counter += 1
        if counter > r:
          latencies.append((end_time - start_time) / 1e9)
    progress.stop()
    avg_latency = sum(latencies) / len(latencies)
    return avg_latency


def test_thruput(sdk: SDK, task: str, size_kb: int, reps: int, warmup=0.1) -> float:
    """
    Measure the thruput
    
    """
    if task == "uload":
        fn = lambda: sdk.uload(size_kb)
    elif task == "dload":
        fn = lambda: sdk.dload(size_kb)
    r = int(reps * warmup)
    counter = 0
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task("counter", total=reps)
    while counter != reps:
        if counter == r:
            start_time = time.time_ns()
        fn()
        progress.update(prog_bar, advance=1)
        counter += 1
    end_time = time.time_ns()
    progress.stop()
    total_size_kb = (size_kb * reps)
    thruput = total_size_kb / 1024 / (end_time - start_time) * 1e9
    return thruput

def test_load(sdk: SDK, size_kb: int, rate: int):
    """
    Load the system
    
    """
    while True:
        fn = lambda: random.choice([sdk.uload, sdk.dload])(size_kb)
        start_time = time.time_ns()
        fn()
        end_time = time.time_ns()
        duration = (end_time - start_time) / 1e9
        time.sleep(max(0, 1/rate - duration))

if __name__ == "__main__":

    linspace = [2**i for i in range(16)]
    parser = argparse.ArgumentParser()

    # args for connection
    parser.add_argument("--conn", type=str, choices=["agent", "ssh"], required=True, help="Connection Type")
    parser.add_argument("--api", type=str, help="[Agent] API URL")
    parser.add_argument("--agent", type=str, help="[Agent] Agent ID")
    parser.add_argument("--proxy", type=str, help="[SSH] Proxy Addr (user@hostname)")
    parser.add_argument("--remote", type=str, help="[SSH] Remote Addr (user@hostname)")

    # args for task
    parser.add_argument("--task", type=str, choices=["bench", "load"], required=True, help="Task to perform")
    parser.add_argument("--size", type=int, default=32, help="[Bench] Payload size (KB)")
    parser.add_argument("--reps", type=int, default=100, help="[Bench] Number of repetitions")
    parser.add_argument("--rate", type=int, default=1, help="[Load] Request rate (req/s)")

    args = parser.parse_args()
    sdk: SDK

    if args.conn == "agent":
        assert args.api is not None
        assert args.agent is not None
        sdk = AgentSDK(args.api, args.agent)
    elif args.conn == "ssh":
        assert args.proxy is not None
        assert args.remote is not None
        sdk = SSHSDK(args.proxy, args.remote)
    else:
        raise ValueError("Invalid connection type. Choose 'agent' or 'ssh'.")

    sdk.setup(linspace)
    if args.task == "bench":
        latency_ul = test_latency(sdk, "uload", args.size, args.reps)
        print(f"Latency (U): {latency_ul:.6f} s")
        latency_dl = test_latency(sdk, "dload", args.size, args.reps)
        print(f"Latency (D): {latency_dl:.6f} s")
        thruput_ul = test_thruput(sdk, "uload", args.size, args.reps)
        print(f"Thruput (U): {thruput_ul:.6f} MB/s")
        thruput_dl = test_thruput(sdk, "dload", args.size, args.reps)
        print(f"Thruput (D): {thruput_dl:.6f} MB/s")
    elif args.task == "load":
        test_load(sdk, args.size, args.rate)
    else:
        raise ValueError("Invalid task. Choose 'bench' or 'load'.")
