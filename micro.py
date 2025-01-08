#!/usr/bin/env python3

# MICRO BENCHMARKS

import time
import argparse
from sdk import AgentSDK, SSHSDK, SDK

def test_latency(sdk: SDK, task: str, size_kb: int, reps: int) -> float:
    """
    Measure the round-trip latency
    
    """
    if task == "uload":
        fn = lambda: sdk.uload(size_kb)
    elif task == "dload":
        fn = lambda: sdk.dload(size_kb)
    latencies = []
    while len(latencies) != reps:
        start_time = time.time_ns()
        fn()
        end_time = time.time_ns()
        latencies.append((end_time - start_time) / 1e9)
    avg_latency = sum(latencies) / len(latencies)
    return avg_latency


def test_thruput(sdk: SDK, task: str, size_kb: int, reps: int) -> float:
    """
    Measure the thruput
    
    """
    if task == "uload":
        fn = lambda: sdk.uload(size_kb)
    elif task == "dload":
        fn = lambda: sdk.dload(size_kb)
    start_time = time.time_ns()
    counter = 0
    while counter != reps:
        fn()
        counter += 1
    end_time = time.time_ns()
    total_size_kb = (size_kb * reps)
    thruput = total_size_kb / 1024 / (end_time - start_time) * 1e9
    return thruput

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

    if args.task == "bench":
        sdk.setup(linspace)
        latency_ul = test_latency(sdk, "uload", args.size, args.reps)
        print(f"Latency (U): {latency_ul:.6f} s")
        latency_dl = test_latency(sdk, "dload", args.size, args.reps)
        print(f"Latency (D): {latency_dl:.6f} s")
        thruput_ul = test_thruput(sdk, "uload", args.size, args.reps)
        print(f"Thruput (U): {thruput_ul:.6f} MB/s")
        thruput_dl = test_thruput(sdk, "dload", args.size, args.reps)
        print(f"Thruput (D): {thruput_dl:.6f} MB/s")
    elif args.task == "load":
        pass
    else:
        raise ValueError("Invalid task. Choose 'bench' or 'load'.")
