#!/usr/bin/env python3

# MICRO BENCHMARKS

import argparse
import json
import random
import time

import rich.progress

from sdk import SDK, SSHSDK, AgentSDK, gRPCSDK


def test_latency(sdk: SDK, task: str, size_kb: int, reps: int, warmup=0.1) -> float:
    """
    Measure the round-trip latency
    
    """
    if task == "UL":
        fn = lambda: sdk.uload(size_kb)
    elif task == "DL":
        fn = lambda: sdk.dload(size_kb)
    latencies = []
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task(f"Latency ({task}) ⏸", total=reps)
    r = int(reps * warmup)
    counter = 0
    sdk.setup(size_kb)
    while counter != reps:
        start_time = time.time_ns()
        fn()
        end_time = time.time_ns()
        progress.update(prog_bar, advance=1)
        counter += 1
        if counter == r:
            progress.update(prog_bar, description=f"Latency ({task}) ⏵")
        if counter > r:
          latencies.append((end_time - start_time) / 1e9)
    avg_latency = sum(latencies) / len(latencies)
    sdk.teardown()
    progress.update(prog_bar, description=f"Latency ({task}): {avg_latency:.6f} s   ")
    progress.stop()
    return avg_latency


def test_thruput(sdk: SDK, task: str, size_kb: int, reps: int, warmup=0.1) -> float:
    """
    Measure the thruput
    
    """
    if task == "UL":
        fn = lambda: sdk.uload(size_kb)
    elif task == "DL":
        fn = lambda: sdk.dload(size_kb)
    r = int(reps * warmup)
    counter = 0
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task(f"Thruput ({task}) ⏸", total=reps)
    sdk.setup(size_kb)
    while counter != reps:
        if counter == r:
            progress.update(prog_bar, description=f"Thruput ({task}) ⏵")
            start_time = time.time_ns()
        fn()
        progress.update(prog_bar, advance=1)
        counter += 1
    end_time = time.time_ns()
    total_size_kb = (size_kb * reps)
    thruput = total_size_kb / 1024 / (end_time - start_time) * 1e9
    sdk.teardown()
    progress.update(prog_bar, description=f"Thruput ({task}): {thruput:.6f} MB/s")
    progress.stop()
    return thruput


def test_command_execution(sdk: SDK, command: str, reps: int, warmup=0.1) -> float:
    """
    Measure the execution latency of a remote command.
    """
    latencies = []
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task("Command Execution Latency", total=reps)
    r = int(reps * warmup)
    counter = 0

    while counter != reps:
        start_time = time.time_ns()
        stdout, stderr = sdk.exec(command, b"")
        end_time = time.time_ns()
        progress.update(prog_bar, advance=1)
        counter += 1

        if counter >= r:
            latencies.append((end_time - start_time) / 1e9)

    avg_latency = sum(latencies) / len(latencies)
    progress.stop()
    return avg_latency


def test_load(sdk: SDK, command: str, rate: int):
    """
    Load the system by executing a specified command with Gaussian-distributed inter-request intervals.
    """
    progress = rich.progress.Progress()
    progress.start()
    prog_bar = progress.add_task("Load", total=None)

    mu = 1 / rate
    sigma = mu * 0.1

    try:
        while True:
            start_time = time.time_ns()
            stdout, stderr = sdk.exec(command, b"")
            progress.update(prog_bar, advance=1)
            end_time = time.time_ns()
            duration = (end_time - start_time) / 1e9
            sleep_time = max(0, random.gauss(mu, sigma) - duration)
            time.sleep(sleep_time)
    except KeyboardInterrupt:
        pass
    finally:
        progress.stop()


if __name__ == "__main__":
    
    parser = argparse.ArgumentParser()
    
    # args for connection
    parser.add_argument("--conn", type=str, choices=["agent", "ssh", "grpc"], required=True, help="Connection Type")
    parser.add_argument("--api", type=str, help="[Agent] API URL")
    parser.add_argument("--agent", type=str, help="[Agent] Agent ID")
    parser.add_argument("--proxy", type=str, help="[SSH] Proxy Addr (user@hostname)")
    parser.add_argument("--remote", type=str, help="[SSH] Remote Addr (user@hostname)")
    parser.add_argument("--cli", type=str, help="[gRPC] CLI Path")
    parser.add_argument("--sock", type=str, help="[gRPC] Sock Path")
    parser.add_argument("--peer", type=str, help="[gRPC] Remote ID")
    
    # args for task
    parser.add_argument("--task", type=str, choices=["bench", "load"], required=True, help="Task to perform")
    parser.add_argument("--cmd", type=str, help="[Load] Command to execute during load testing")
    parser.add_argument("--reps", type=int, default=100, help="[Bench] Number of repetitions")
    parser.add_argument("--rate", type=int, default=1, help="[Load] Request rate (req/s)")
    parser.add_argument("--cmd", type=str, help="Command to execute for CMD task")

    # args to save results
    parser.add_argument("--dest", type=str, help="Destination to save results")
    
    # parse args
    args = parser.parse_args()
    sdk: SDK
    
    # logic
    if args.conn == "agent":
        assert args.api is not None
        assert args.agent is not None
        sdk = AgentSDK(args.api, args.agent)
    elif args.conn == "ssh":
        assert args.proxy is not None
        assert args.remote is not None
        sdk = SSHSDK(args.proxy, args.remote)
    elif args.conn == "grpc":
        assert args.cli is not None
        assert args.sock is not None
        assert args.peer is not None
        sdk = gRPCSDK(args.cli, args.sock, args.peer)
    else:
        raise ValueError("Invalid connection type. Choices=['agent', 'ssh', 'grpc']")
    
    if args.task == "bench":
        # run tests
        print("Running Benchmarks...")
        latency_ul = test_latency(sdk, "UL", args.size, args.reps)
        latency_dl = test_latency(sdk, "DL", args.size, args.reps)
        thruput_ul = test_thruput(sdk, "UL", args.size, args.reps)
        thruput_dl = test_thruput(sdk, "DL", args.size, args.reps)
        # save summary
        result = dict(
            size_kb=args.size,
            latency_ul=latency_ul,
            latency_dl=latency_dl,
            thruput_ul=thruput_ul,
            thruput_dl=thruput_dl
        )
        line = json.dumps(result)
        with open(args.dest, "a") as f:
            f.write(line + "\n")
        print("latency_ul,latency_dl,thruput_ul,thruput_dl")
        print(f"{latency_ul:.6f},{latency_dl:.6f},{thruput_ul:.6f},{thruput_dl:.6f}")
    elif args.task == "load":
        print("Creating Load...")
        test_load(sdk, args.cmd, args.rate)
    elif args.task == "cmd":
        if not args.cmd:
            raise ValueError("Command must be provided for CMD task")
        latency_cmd = test_command_execution(sdk, args.cmd, args.reps)
        result = dict(command=args.cmd, latency_cmd=latency_cmd)
        with open(args.dest, "a") as f:
            f.write(json.dumps(result) + "\n")
        print(f"Command Execution Latency: {latency_cmd:.6f} s")
    else:
        raise ValueError("Invalid task. Choose 'bench' or 'load'.")
