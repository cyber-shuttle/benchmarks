#!/usr/bin/env python3

import argparse
import matplotlib.pyplot as plt
import pandas as pd


def plot_results(input_file, output_file):
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
    plt.xticks(df['num_agents'][::5])
    plt.xlabel('Number of Agents')
    plt.ylabel('Latency (milliseconds)')
    plt.title('Latency Metrics vs. Number of Agents')
    plt.legend()
    plt.grid(True)

    plt.savefig(output_file)
    plt.close()

    print(f"Plot saved to {output_file}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Plot Latency Metrics")
    parser.add_argument("--input_file", type=str, required=True, help="Input JSONL file containing results")
    parser.add_argument("--output_file", type=str, required=True, help="Output PNG file for the plot")
    args = parser.parse_args()

    plot_results(args.input_file, args.output_file)
