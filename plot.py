import pandas as pd
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
