#!/usr/bin/env python3

import argparse
import matplotlib.pyplot as plt
import numpy as np
import os
import pandas as pd


def plot_latency_results(input_file, output_dir):
    """
    Plot the latency statistics against the number of agents
    """
    try:
        df = pd.read_json(input_file, lines=True)
    except ValueError as e:
        print(f"Error reading the input file {input_file}: {e}")
        return

    # Ensure the DataFrame has the required columns
    required_columns = ['num_agents', 'mean_latency', 'median_latency', 'p90_latency', 'p95_latency']
    if not all(column in df.columns for column in required_columns):
        print(f"Input file is missing required columns: {required_columns}")
        return

    plt.figure(figsize=(10, 6))
    plt.plot(df['num_agents'], df['mean_latency'], marker='o', label='Mean Latency')
    plt.plot(df['num_agents'], df['median_latency'], marker='o', label='Median Latency')
    plt.plot(df['num_agents'], df['p90_latency'], marker='o', label='90th Percentile Latency')
    plt.plot(df['num_agents'], df['p95_latency'], marker='o', label='95th Percentile Latency')

    # plt.xticks(df['num_agents'])
    plt.xticks(df['num_agents'][::10])
    plt.xlabel('Number of Clients')
    plt.ylabel('Latency (milliseconds)')
    plt.title('Latency Metrics vs. Number of Clients')
    plt.legend()
    plt.grid(True)

    output_file = os.path.join(output_dir, "latency_results.png")
    plt.savefig(output_file)
    plt.close()

    print(f"Latency results plot saved to {output_file}")


def plot_latency_distribution(input_file, num_agents, output_dir):
    """
    Plot the latency distribution for a specific number of agents
    """
    try:
        df = pd.read_json(input_file, lines=True)
    except ValueError as e:
        print(f"Error reading the input file {input_file}: {e}")
        return

    # Filter data for the specified number of agents
    filtered_df = df[df['num_agents'] == num_agents]
    if filtered_df.empty:
        print(f"No data found for {num_agents} agents.")
        return

    latencies = []
    if 'latencies' in filtered_df.columns:
        latencies = filtered_df.iloc[0]['latencies']
    else:
        print(f"Latencies data missing for {num_agents} agents.")
        return

    # Calculate key metrics
    mean_latency = np.mean(latencies)
    median_latency = np.median(latencies)
    p90_latency = np.percentile(latencies, 90)
    p95_latency = np.percentile(latencies, 95)

    plt.figure(figsize=(10, 6))
    plt.hist(latencies, bins=20, alpha=0.7, color='skyblue', edgecolor='black', label='Latencies')

    plt.axvline(mean_latency, color='blue', linestyle='--', label=f'Mean ({mean_latency:.2f} ms)')
    plt.axvline(median_latency, color='orange', linestyle='--', label=f'Median ({median_latency:.2f} ms)')
    plt.axvline(p90_latency, color='green', linestyle='--', label=f'90th Percentile ({p90_latency:.2f} ms)')
    plt.axvline(p95_latency, color='red', linestyle='--', label=f'95th Percentile ({p95_latency:.2f} ms)')

    plt.title(f'Latency Distribution for {num_agents} Clients')
    plt.xlabel('Latency (ms)')
    plt.ylabel('Frequency')
    plt.legend()
    plt.grid(True)

    output_file = os.path.join(output_dir, f"latency_distribution_{num_agents}_agents.png")
    plt.savefig(output_file)
    plt.close()
    print(f"Latency distribution plot for {num_agents} agents saved to {output_file}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Plot Latency Metrics")
    parser.add_argument("--input_file", type=str, required=True, help="Input JSONL file containing results")
    parser.add_argument("--output", type=str, required=True, help="Output location for the plots")
    parser.add_argument("--num_agents", type=int, required=True,
                        help="Number of agents to analyze for latency distribution")
    args = parser.parse_args()

    os.makedirs(args.output, exist_ok=True)

    plot_latency_results(args.input_file, args.output)
    plot_latency_distribution(args.input_file, args.num_agents, args.output)
