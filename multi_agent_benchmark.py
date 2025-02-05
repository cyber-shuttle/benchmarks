#!/usr/bin/env python3

import argparse
import json
import matplotlib
import numpy as np
import os
import pandas as pd
import subprocess
import time

matplotlib.use('Agg')
import matplotlib.pyplot as plt
from concurrent.futures import ThreadPoolExecutor, as_completed

from sdk import gRPCSDK


def start_agent(agent_id, sock_file, server_address):
    """
    Start an agent process.
    """
    agent_cmd = [
        "./agent_amd64",
        "-r", server_address,
        "-i", agent_id,
        "-s", sock_file
    ]
    return subprocess.Popen(agent_cmd)


def run_micro(agent_id, sock_file, peer_id, command, duration, num_executions):
    """
    Run the micro benchmark using the gRPC connection
    """
    sdk = gRPCSDK(cli="/home/ubuntu/benchmarks/grpcsh_amd64", sock=sock_file, peer=peer_id)
    latencies = []
    interval = duration / num_executions

    for _ in range(num_executions):
        req_start = time.time_ns()
        stdout, stderr = sdk.exec(command, b"")
        req_end = time.time_ns()
        latencies.append((req_end - req_start) / 1e6)  # To milliseconds
        time.sleep(interval)

    return latencies


def aggregate_statistics(latencies):
    """
    Calculate mean, median, 90th percentile, and 95th percentile latencies
    """
    if not latencies:
        return {
            "mean_latency": 0,
            "median_latency": 0,
            "p90_latency": 0,
            "p95_latency": 0
        }

    latencies = np.array(latencies)
    mean_latency = np.mean(latencies)
    median_latency = np.median(latencies)
    p90_latency = np.percentile(latencies, 90)
    p95_latency = np.percentile(latencies, 95)

    return {
        "mean_latency": mean_latency,
        "median_latency": median_latency,
        "p90_latency": p90_latency,
        "p95_latency": p95_latency
    }


def plot_results(results, dest):
    """
    Plot the latency statistics against the number of agents
    """
    df = pd.DataFrame(results)

    plt.figure(figsize=(10, 6))
    plt.plot(df['num_agents'], df['mean_latency'], marker='o', label='Mean Latency')
    plt.plot(df['num_agents'], df['median_latency'], marker='o', label='Median Latency')
    plt.plot(df['num_agents'], df['p90_latency'], marker='o', label='90th Percentile Latency')
    plt.plot(df['num_agents'], df['p95_latency'], marker='o', label='95th Percentile Latency')

    plt.xticks(df['num_agents'])
    plt.xlabel('Number of Agents')
    plt.ylabel('Latency (milliseconds)')
    plt.title('Latency Metrics vs. Number of Agents')
    plt.legend()
    plt.grid(True)

    plot_file = os.path.splitext(dest)[0] + "_latency_metrics_plot.png"
    plt.savefig(plot_file)
    plt.close()

    print(f"Plot saved to {plot_file}")


def main(server_address, command, duration, max_agents, num_executions, dest):
    agents = []
    results = []
    os.makedirs(os.path.dirname(dest), exist_ok=True)

    # Start all agents without executing commands
    for i in range(1, max_agents + 1):
        agent_id = f"client_agent_{i}"
        sock_file = f"/home/ubuntu/{agent_id}.sock"
        agent_process = start_agent(agent_id, sock_file, server_address)
        agents.append(agent_process)
        # Time for the agent to initialize
        time.sleep(2)

    # Sequentially start command execution for each agent
    with ThreadPoolExecutor(max_workers=max_agents) as executor:
        for i in range(1, max_agents + 1):
            futures = []
            for j in range(1, i + 1):
                agent_id = f"client_agent_{j}"
                peer_id = f"agent_{j}"
                sock_file = f"/home/ubuntu/{agent_id}.sock"
                futures.append(
                    executor.submit(run_micro, agent_id, sock_file, peer_id, command, duration, num_executions))

            all_latencies = []
            for future in as_completed(futures):
                all_latencies.extend(future.result())

            stats = aggregate_statistics(all_latencies)
            stats["num_agents"] = i
            results.append(stats)

            # Save intermediate results
            with open(dest, "a") as f:
                f.write(json.dumps(stats) + "\n")

            print(f"Completed benchmark with {i} agent(s): {stats}")

    for agent in agents:
        agent.terminate()
        agent.wait()
    print(f"Terminated the agents")

    plot_results(results, dest)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Multi-Agent Benchmarking Script")
    parser.add_argument("--server_address", type=str, required=True, help="Server address (e.g., 3.15.162.26:50051)")
    parser.add_argument("--command", type=str, required=True, help="Command to execute")
    parser.add_argument("--duration", type=int, default=120,
                        help="Duration for each agent to run the command (in seconds)")
    parser.add_argument("--max_agents", type=int, default=10, help="Maximum number of agents to simulate")
    parser.add_argument("--num_executions", type=int, default=5,
                        help="Number of command executions per agent per duration")
    parser.add_argument("--dest", type=str, required=True,
                        help="Destination file to save results (e.g., /home/ubuntu/results.jsonl)")

    args = parser.parse_args()
    main(args.server_address, args.command, args.duration, args.max_agents, args.num_executions, args.dest)
