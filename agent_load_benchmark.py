#!/usr/bin/env python3

import argparse
import json
import numpy as np
import os
import random
import subprocess
import time

from sdk import gRPCSDK


def start_agent(agent_id, sock_file, server_address):
    """
    Start a single agent process connected to the remote system.
    """
    agent_cmd = [
        "./agent_amd64",
        "-r", server_address,
        "-i", agent_id,
        "-s", sock_file
    ]
    return subprocess.Popen(agent_cmd)


def gaussian_load(server_address, command, request_rate, duration, dest):
    """
    Generate load using Gaussian-distributed inter-request intervals with a dedicated agent
    """
    agent_id = "client_agent_gauss_load"
    sock_file = f"/home/ubuntu/{agent_id}.sock"

    agent_process = start_agent(agent_id, sock_file, server_address)
    sdk = gRPCSDK(cli="/home/ubuntu/benchmarks/grpcsh_amd64", sock=sock_file, peer="remote_agent")

    inter_request_intervals = []
    mu = 1 / request_rate
    sigma = mu * 0.1

    start_time = time.time()
    try:
        while (time.time() - start_time) < duration:
            sdk.exec(command, b"")

            sleep_time = max(0, random.gauss(mu, sigma))
            time.sleep(sleep_time)
            inter_request_intervals.append(sleep_time)

            with open(dest, "a") as f:
                f.write(json.dumps({
                    "request_rate": request_rate,
                    "inter_request_interval": sleep_time
                }) + "\n")

    except Exception as e:
        print(f"Error during Gaussian load generation: {e}")

    finally:
        agent_process.terminate()
        agent_process.wait()
        print(f"Gaussian load agent terminated after {duration} seconds.")

    print(f"Load generation for {request_rate} req/s completed.")


def run_benchmark(server_address, command, duration, dest, request_rate):
    """
    The benchmark agent runs at a **fixed 1 request per second.** under the load condition
    """
    agent_id = "client_agent_benchmark"
    sock_file = f"/home/ubuntu/{agent_id}.sock"

    agent_process = start_agent(agent_id, sock_file, server_address)
    sdk = gRPCSDK(cli="/home/ubuntu/benchmarks/grpcsh_amd64", sock=sock_file, peer="remote_agent")
    latencies = []

    start_time = time.time()
    try:
        while (time.time() - start_time) < duration:
            req_start = time.time_ns()
            sdk.exec(command, b"")
            req_end = time.time_ns()

            latency = (req_end - req_start) / 1e6  # To milliseconds
            latencies.append(latency)

            # Ensure a constant **1 request per second**
            time.sleep(1)

    except Exception as e:
        print(f"Error during benchmarking: {e}")

    finally:
        agent_process.terminate()
        agent_process.wait()

    with open(dest, "a") as f:
        f.write(json.dumps({
            "load_request_rate": request_rate,
            "mean_latency": np.mean(latencies) if latencies else 0,
            "median_latency": np.median(latencies) if latencies else 0,
            "p90_latency": np.percentile(latencies, 90) if latencies else 0,
            "p95_latency": np.percentile(latencies, 95) if latencies else 0,
            "latencies": latencies
        }) + "\n")

    print(f"Benchmark completed and saved to {dest}.")


def warmup(server_address, command, duration):
    """
    Perform a warm-up before benchmarking
    """
    agent_id = "client_agent_warmup"
    sock_file = f"/home/ubuntu/{agent_id}.sock"

    agent_process = start_agent(agent_id, sock_file, server_address)
    sdk = gRPCSDK(cli="/home/ubuntu/benchmarks/grpcsh_amd64", sock=sock_file, peer="remote_agent")

    start_time = time.time()
    try:
        while (time.time() - start_time) < duration:
            sdk.exec(command, b"")
            time.sleep(1)  # Constant warm-up rate (1 req/s)
        print("Warm-up phase completed.")
    finally:
        agent_process.terminate()
        agent_process.wait()


def main(server_address, command, duration, max_request_rate, request_step, dest):
    os.makedirs(os.path.dirname(dest), exist_ok=True)

    print(f"Starting warm-up phase...")
    warmup(server_address, command, duration)
    print(f"Warm-up phase completed. Starting benchmarks...")

    for request_rate in range(1, max_request_rate + 1, request_step):
        print(f"Running Gaussian load at {request_rate} req/s...")
        gaussian_load(server_address, command, request_rate, duration, dest + "_gaussian.jsonl")

        print(f"Running benchmark under {request_rate} req/s background load...")
        run_benchmark(server_address, command, duration, dest + "_benchmark.jsonl", request_rate)

    print(f"Benchmarking completed. Results saved to {dest}_gaussian.jsonl and {dest}_benchmark.jsonl")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Gaussian Load Benchmarking Script")
    parser.add_argument("--server_address", type=str, required=True, help="Server address (e.g., 3.15.162.26:50051)")
    parser.add_argument("--command", type=str, required=True, help="Command to execute")
    parser.add_argument("--duration", type=int, default=60, help="Duration for each request rate (in seconds)")
    parser.add_argument("--max_request_rate", type=int, default=100, help="Maximum request rate (req/s)")
    parser.add_argument("--request_step", type=int, default=10, help="Step size for increasing request rate")
    parser.add_argument("--dest", type=str, required=True, help="Destination file prefix (e.g., /home/ubuntu/results)")

    args = parser.parse_args()
    main(args.server_address, args.command, args.duration, args.max_request_rate, args.request_step, args.dest)
