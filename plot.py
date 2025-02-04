import pandas as pd
import numpy as np
from matplotlib import pyplot as plt
import json

def read_to_df(file_map: dict[str, str]) -> pd.DataFrame:
  dfs = []
  for typ, file_path in file_map.items():
    with open(file_path, 'r') as file:
      data = [json.loads(line) for line in file]
      df = pd.DataFrame(data)
      df["type"] = typ
      dfs.append(df.set_index("size_kb"))
  return pd.concat(dfs).reset_index()


def plot_theoretical_gaussian_histogram(R, num_samples=1000):
  # Simulate Gaussian inter-request intervals
  mu = 1 / R
  sigma = mu * 0.1
  samples = np.random.normal(mu, sigma, num_samples)

  plt.hist(samples, bins=30, alpha=0.75, edgecolor="black")
  plt.xlabel("Inter-Request Interval (seconds)")
  plt.ylabel("Frequency")
  plt.title(f"Gaussian Distributed Inter-Request Intervals (R={R} req/sec)")
  plt.grid(True)
  plt.show()


def plot_inter_request_intervals(file_path):
  """Generate and save plots for inter-request intervals."""
  intervals = []

  with open(file_path, "r") as f:
    for line in f:
      data = json.loads(line)
      intervals.append(data["inter_request_interval"])

  #Histogram: Distribution of Inter-Request Intervals
  plt.figure(figsize=(8,5))
  plt.hist(intervals, bins=20, edgecolor="black", alpha=0.75)
  plt.xlabel("Inter-Request Interval (seconds)")
  plt.ylabel("Frequency")
  plt.title("Distribution of Inter-Request Intervals")
  plt.grid(True)
  plt.savefig("inter_request_interval_histogram.png")
  plt.close()

  #Line Plot: How Intervals Change Over Time
  plt.figure(figsize=(8,5))
  plt.plot(intervals, marker='o', linestyle='-', markersize=3)
  plt.xlabel("Request Number")
  plt.ylabel("Inter-Request Interval (seconds)")
  plt.title("Inter-Request Intervals Over Time")
  plt.grid(True)
  plt.savefig("inter_request_interval_time_series.png")
  plt.close()

  print("Plots saved as 'inter_request_interval_histogram.png' and 'inter_request_interval_time_series.png'")


if __name__ == "__main__":

  idx = "size_kb"
  col = "type"
  val = ["latency_ul", "latency_dl", "thruput_ul", "thruput_dl"]
  df = read_to_df({
    "ssh": "idle_ssh.jsonl",
    "grpc": "idle_grpc.jsonl"
  }).pivot(index=idx, columns=col, values=val)

  # Plot Latency
  fig, ax1 = plt.subplots(figsize=(10, 6))
  df[['latency_dl', 'latency_ul']].plot(ax=ax1, marker='o')
  ax1.set_title('Latency Comparison: gRPC vs SSH')
  ax1.set_xlabel('Size (KB)')
  ax1.set_ylabel('Latency (ms)')
  ax1.set_xscale('log')
  ax1.grid(True)
  plt.legend(['latency_dl_grpc', 'latency_ul_grpc', 'latency_dl_ssh', 'latency_ul_ssh'])

  # Plot Throughput
  fig, ax2 = plt.subplots(figsize=(10, 6))
  df[['thruput_dl', 'thruput_ul']].plot(ax=ax2, marker='o')
  ax2.set_title('Throughput Comparison: gRPC vs SSH')
  ax2.set_xlabel('Size (KB)')
  ax2.set_ylabel('Throughput (MB/s)')
  ax2.set_xscale('log')
  ax2.grid(True)
  plt.legend(['thruput_dl_grpc', 'thruput_ul_grpc', 'thruput_dl_ssh', 'thruput_ul_ssh'])
  
  
  plt.show()
