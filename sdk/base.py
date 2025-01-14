from abc import ABC, abstractmethod


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