#!/usr/bin/env python3

import argparse
import matplotlib.pyplot as plt
import numpy as np
import os
import pandas as pd


def plot_latency_results(benchmark_file, output_dir):
    """
    Plot the latency statistics against the Gaussian load request rate
    """
    try:
        df = pd.read_json(benchmark_file, lines=True)
    except ValueError as e:
        print(f"Error reading the benchmark file {benchmark_file}: {e}")
        return

    required_columns = ['load_request_rate', 'mean_latency', 'median_latency', 'p90_latency', 'p95_latency']
    if not all(column in df.columns for column in required_columns):
        print(f"Benchmark file is missing required columns: {required_columns}")
        return

    plt.figure(figsize=(10, 6))
    plt.plot(df['load_request_rate'], df['mean_latency'], marker='o', label='Mean Latency')
    plt.plot(df['load_request_rate'], df['median_latency'], marker='o', label='Median Latency')
    plt.plot(df['load_request_rate'], df['p90_latency'], marker='o', label='90th Percentile Latency')
    plt.plot(df['load_request_rate'], df['p95_latency'], marker='o', label='95th Percentile Latency')

    plt.xticks(df['load_request_rate'][::2])  # Adjust tick frequency if needed
    plt.xlabel('Gaussian Load Request Rate (req/s)')
    plt.ylabel('Latency (milliseconds)')
    plt.title('Latency Metrics vs. Gaussian Load Request Rate')
    plt.legend()
    plt.grid(True)

    output_file = os.path.join(output_dir, "latency_vs_gaussian_load.pdf")
    plt.savefig(output_file)
    plt.close()

    print(f"Latency results plot saved to {output_file}")


def plot_latency_distribution(benchmark_file, request_rate, output_dir):
    """
    Plot the latency distribution for a specific Gaussian request rate
    """
    try:
        df = pd.read_json(benchmark_file, lines=True)
    except ValueError as e:
        print(f"Error reading the benchmark file {benchmark_file}: {e}")
        return

    # Filter data for the selected request rate
    filtered_df = df[df['load_request_rate'] == request_rate]
    if filtered_df.empty:
        print(f"No data found for request rate {request_rate} req/s.")
        return

    latencies = []
    if 'latencies' in filtered_df.columns:
        latencies = filtered_df.iloc[0]['latencies']
    else:
        print(f"Latencies data missing for request rate {request_rate} req/s.")
        return

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

    plt.title(f'Latency Distribution for Load = {request_rate} req/s')
    plt.xlabel('Latency (ms)')
    plt.ylabel('Frequency')
    plt.legend()
    plt.grid(True)

    output_file = os.path.join(output_dir, f"latency_distribution_{request_rate}_reqs.pdf")
    plt.savefig(output_file)
    plt.close()
    print(f"Latency distribution plot for {request_rate} req/s saved to {output_file}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Plot Latency Metrics for Gaussian Load Benchmark")
    parser.add_argument("--benchmark_file", type=str, required=True, help="Input JSONL file containing benchmark results")
    parser.add_argument("--output", type=str, required=True, help="Output directory for the plots")
    parser.add_argument("--request_rate", type=int, required=True, help="Request rate to analyze for latency distribution")

    args = parser.parse_args()
    os.makedirs(args.output, exist_ok=True)

    plot_latency_results(args.benchmark_file, args.output)
    plot_latency_distribution(args.benchmark_file, args.request_rate, args.output)
