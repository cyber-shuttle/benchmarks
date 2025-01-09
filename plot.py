import pandas as pd
from matplotlib import pyplot as plt
import json

if __name__ == "__main__":
  file_path = 'home_nw_idle.jsonl'

  # Open the file and read it line by line
  with open(file_path, 'r') as file:
      # Load each JSON line into a list of dictionaries
      data = [json.loads(line) for line in file]

  # Create a DataFrame from the list of dictionaries
  df = pd.DataFrame(data)

  # Display the first few rows of the DataFrame
  fig = plt.figure()
  ax = fig.add_subplot(111)
  ax.plot(df['size_kb'], df['latency_ul'], label='latency_ul')
  ax.plot(df['size_kb'], df['latency_dl'], label='latency_dl')
  ax.plot(df['size_kb'], df['thruput_ul'], label='thruput_ul')
  ax.plot(df['size_kb'], df['thruput_dl'], label='thruput_dl')
  plt.suptitle("Home Network - Idle")
  plt.legend()
  plt.show()