import base64
import os
import time

import requests

from .base import SDK


class AgentSDK(SDK):
  """
  ============================================================================
  Proposed Method
  ============================================================================
                                   -----------------------------------
  * <sdk@client> -> <api@proxy> <- | <agent@remote> -> <bash@remote> |
                                   -----------------------------------
  - api@proxy spawns agent@remote
  - agent@remote connects to api@proxy
  - agent@remote opens bash
  ----------------------------------------------------------------------------
  * sdk@client requests bash@remote [thru api@proxy] [from agent@remote]
  * @client has credentials for ONLY @proxy
  * @proxy has credentials for ONLY @remote
  ----------------------------------------------------------------------------
  """

  # TODO next step - setup/teardown create persistent connection to agent

  api_url: str
  agent_id: str

  def __init__(self, api_url: str, agent_id: str) -> None:
    self.api_url = api_url
    self.agent_id = agent_id

  def __exec_command__(self, cmd: str, wait_ms: int = 100) -> bytes:
    # send request and get execution id
    res = requests.post(f"{self.api_url}/executecommandrequest", json={
      "agentId": self.agent_id,
      "workingDir": ".",
      "arguments": ["sh", "-c", cmd]
    })
    assert res.ok == True
    data = res.json()
    assert data["error"] is None
    exc_id = data["executionId"]
    # wait for execution to complete (polling until available)
    data = None
    while True:
      res = requests.get(f"{self.api_url}/executecommandresponse/{exc_id}")
      assert res.ok == True
      data = res.json()
      if data["available"]:
        break
      time.sleep(wait_ms / 1000)
    return str(data["responseString"]).encode()

  def setup(self, size_kb: int) -> None:
    res = requests.get(f"{self.api_url}/{self.agent_id}")
    assert res.status_code == 202
    assert res.json()["agentUp"] == True
    res = self.__exec_command__(f"base64 < /dev/urandom | head -c {size_kb * 1024} > data_{size_kb}k")

  def teardown(self) -> None:
    # nothing to teardown for now
    pass

  def uload(self, size_kb: int) -> bytes:
    data = base64.b64encode(os.urandom(size_kb * 1024))[:size_kb * 1024]
    res = self.__exec_command__(f"echo {data} > data_{size_kb}k")
    return data

  def dload(self, size_kb: int) -> bytes:
    data = self.__exec_command__(f"cat data_{size_kb}k")
    return data

  def exec(self, cmd: str, input: bytes) -> tuple[bytes, bytes]:
    res = self.__exec_command__(cmd)
    # doing nothing with the input bytes for now (no data channel yet)
    return b"", b""

