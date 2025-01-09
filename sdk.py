from abc import ABC, abstractmethod
import os
import paramiko
import requests


class TestCase(ABC):

  @abstractmethod
  def setup(self, size_kb: int) -> None: ...

  @abstractmethod
  def teardown(self) -> None: ...


class SDK(TestCase):

  @abstractmethod
  def uload(self, size_kb: int) -> bytes: ...

  @abstractmethod
  def dload(self, size_kb: int) -> bytes: ...

  @abstractmethod
  def exec(self, cmd: str, input: bytes) -> tuple[bytes, bytes]: ...


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
  * sdk@client requests bash@remote from api@proxy [agent@remote is internal]
  * @client has credentials for ONLY @proxy
  * @proxy has credentials for ONLY @remote
  ----------------------------------------------------------------------------
  """

  api_url: str
  agent_id: str

  def __init__(self, api_url: str, agent_id: str) -> None:
    self.api_url = api_url
    self.agent_id = agent_id

  def setup(self, size_kb: int) -> None:
    res = requests.get(f"{self.api_url}/{self.agent_id}")
    assert res.status_code == 202
    assert res.json()["agentUp"] == True

  def teardown(self) -> None:
    pass

  def uload(self, size_kb: int) -> bytes:
    return b""

  def dload(self, size_kb: int) -> bytes:
    return b""

  def exec(self, cmd: str, input: bytes) -> tuple[bytes, bytes]:
    return b"", b""


class SSHSDK(SDK):
  """
  ============================================================================
  Baseline Method
  ============================================================================
                                    ----------------------------------
  * <sdk@client> -> <sshd@proxy> -> | <sshd@remote> -> <bash@remote> |
                                    ----------------------------------
  - sshd@proxy connects to sshd@remote
  - sshd@remote opens bash
  ---------------------------------------------------------------------------
  * sdk@client requests bash@remote from sshd@remote, via sshd@proxy
  * @client has credentials for BOTH @proxy AND @remote
  ---------------------------------------------------------------------------
  """

  proxy_uname: str
  proxy_host: str
  remote_uname: str
  remote_host: str
  proxy_client: paramiko.SSHClient
  proxy_transp: paramiko.Transport
  bridge_channel: paramiko.Channel
  remote_client: paramiko.SSHClient

  def __init__(self, proxy_addr: str, remote_addr: str) -> None:
    (self.proxy_uname, self.proxy_host) = proxy_addr.split("@")
    (self.remote_uname, self.remote_host) = remote_addr.split("@")

  def setup(self, size_kb: int) -> None:
    # Connect to proxy
    self.proxy_client = paramiko.SSHClient()
    self.proxy_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    self.proxy_client.connect(username=self.proxy_uname, hostname=self.proxy_host)

    # Create a bridge channel from proxy to remote
    t = self.proxy_client.get_transport()
    assert t is not None
    self.proxy_transp = t
    self.bridge_channel = self.proxy_transp.open_channel("direct-tcpip", (self.remote_host, 22), ("", 0))

    # Connect to remote through tunnel
    self.remote_client = paramiko.SSHClient()
    self.remote_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    self.remote_client.connect(username=self.remote_uname, hostname=self.remote_host, sock=self.bridge_channel)

    # create data file
    (stdin, stdout, stderr) = self.remote_client.exec_command(f"head -c {size_kb * 1024} < /dev/urandom > data_{size_kb}k")

  def teardown(self) -> None:
    self.remote_client.close()
    self.bridge_channel.close()
    self.proxy_client.close()

  def uload(self, size_kb: int) -> None:
    data = os.urandom(size_kb * 1024)
    stdin, stdout, stderr = self.remote_client.exec_command(f"cat > data_{size_kb}k")
    stdin.write(data)
    stdin.flush()

  def dload(self, size_kb: int) -> bytes:
    stdin, stdout, stderr = self.remote_client.exec_command(f"cat data_{size_kb}k")
    data = stdout.read()
    return data

  def exec(self, cmd: str, input: bytes) -> tuple[bytes, bytes]:
    stdin, stdout, stderr = self.remote_client.exec_command(cmd)
    stdin.write(input)
    stdin.flush()
    return stdout.read(), stderr.read()
