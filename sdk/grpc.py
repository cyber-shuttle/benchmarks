import base64
import os
import subprocess

from .base import SDK


class gRPCSDK(SDK):
  """
  ============================================================================
  Proposed Method
  ============================================================================
                                   -----------------------------------
  * <agent@client> -> <router@proxy> <- | <agent@remote> -> <bash@remote> |
                                   -----------------------------------
  - router@proxy spawns agent@remote
  - agent@remote connects to router@proxy
  - agent@remote opens bash
  ----------------------------------------------------------------------------
  * agent@client requests bash@remote [thru router@proxy] [from agent@remote]
  * @client has credentials for ONLY @proxy
  * @proxy has credentials for ONLY @remote
  ----------------------------------------------------------------------------
  """

  cli_args: list[str]

  def __init__(self, cli: str, sock: str, peer: str) -> None:
    self.cli_args = [cli, "-s", sock, "-i", peer, "-c"]

  def setup(self, size_kb: int) -> None:
    p = subprocess.Popen([*self.cli_args, f"base64 < /dev/urandom | head -c {size_kb * 1024} > data_{size_kb}k"])
    p.wait()
    assert p.wait() == 0

  def teardown(self) -> None:
    # nothing to do
    pass

  def uload(self, size_kb: int) -> None:
    data = base64.b64encode(os.urandom(size_kb * 1024))[:size_kb * 1024]
    p = subprocess.Popen([*self.cli_args, f"cat > data_{size_kb}k"], stdin=subprocess.PIPE)
    stdin = p.stdin
    assert stdin is not None
    stdin.write(data)
    stdin.close()
    assert p.wait() == 0

  def dload(self, size_kb: int) -> bytes:
    p = subprocess.Popen([*self.cli_args, f"cat data_{size_kb}k"], stdout=subprocess.PIPE)
    stdout = p.stdout
    assert stdout is not None
    data = stdout.read()
    assert p.wait() == 0
    return data

  def exec(self, cmd: str, input: bytes) -> tuple[bytes, bytes]:
    p = subprocess.Popen([*self.cli_args, cmd],
      stdin=subprocess.PIPE,
      stdout=subprocess.PIPE,
      stderr=subprocess.PIPE
    )
    stdin, stdout, stderr = p.stdin, p.stdout, p.stderr
    assert stdin is not None
    assert stdout is not None
    assert stderr is not None
    stdin.write(input)
    stdin.close()
    out, err = stdout.read(), stderr.read()
    assert p.wait() == 0
    return out, err
