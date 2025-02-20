from typing import Callable
import pydantic as pd
import abc
import seaborn as sns
import matplotlib.pyplot as plt


class Actor(pd.BaseModel, abc.ABC):

    id: str
    t: int = 0

    @abc.abstractmethod
    def forward(self, endt: int) -> None: ...

    @abc.abstractmethod
    def peek(self, endt: int) -> int: ...

    @abc.abstractmethod
    def dump(self) -> dict: ...


class Executor(Actor):

    C: int
    X: list[int]  # stores=tasks, prop=endt
    Q: list[int]  # stores=tasks, prop=durn

    def forward(self, endt) -> None:
        assert self.t <= endt
        self.X.sort()
        # [time=t] move pending -> running
        while (len(self.X) < self.C) and (len(self.Q) > 0):
            self.X.append(self.t + self.Q.pop(0))
            self.X.sort()
        # simulate [time=endt]
        while (len(self.X) > 0) and (self.X[0] <= endt):
            self.t = self.X.pop(0)
            if len(self.Q) > 0:
                self.X.append(self.t + self.Q.pop(0))
                self.X.sort()
        # reach end time
        self.t = endt

    def peek(self, endt) -> int:
        # how far into future it can jump
        return min(endt, endt, *self.X)

    def dump(self) -> dict:
        return {
            "t": self.t,
            "id": self.id,
            "pending": len(self.Q),
            "running": len(self.X),
            "capacity": self.C,
        }


class Scheduler(Actor):

    Q: list[int]  # stores=tasks, prop=durn
    R: list[Executor]
    cost_fn: Callable[..., float]

    @staticmethod
    def linear_cost(
        c: float,  # capacity
        x: float,  # num running tasks
        q: float,  # num pending tasks
    ) -> float:
        return x/c + q/c

    @staticmethod
    def exponential_cost(
        c: float,  # capacity
        x: float,  # num running tasks
        q: float,  # num pending tasks
    ) -> float:
        return x/c * q/c

    def forward(self, endt: int) -> None:
        assert self.t <= endt
        # scheduling logic
        decisions = {}
        while len(self.Q) > 0:
            costs = [
                self.cost_fn(c=Ri.C, x=len(Ri.X), q=len(Ri.Q))
                for Ri in self.R
            ]
            min_val = min(costs)
            min_idx = costs.index(min_val)
            # wait if min cost is too high
            if min_val >= 100:
                break
            self.R[min_idx].Q.append(self.Q.pop(0))
            k = self.R[min_idx].id
            if k not in decisions:
                decisions[k] = 0
            decisions[k] += 1
        # reach end time
        self.t = endt

    def peek(self, endt) -> int:
        # how far into future to jump
        return endt

    def dump(self) -> dict:
        return {
            "t": self.t,
            "id": self.id,
            "pending": len(self.Q),
        }


def simulate(system: list[Actor], start_time: int, end_time: int) -> None:
    current_time = start_time
    snapshots = []
    # simulate the system
    while current_time < end_time:
        next_time = min(end_time, *[actor.peek(end_time) for actor in SYSTEM])
        for actor in SYSTEM:
            snapshots.append(actor.dump())
            actor.forward(next_time)
        current_time = next_time
    # print the dynamics
    import pandas as pd
    df = pd.DataFrame.from_records(snapshots)
    print(df)

    ax = plt.subplot(2, 2, 1)
    sns.lineplot(data=df[(df.id.str.startswith("SCHD"))],
                 x="t", y="pending", hue="id", ax=ax)
    ax = plt.subplot(2, 2, 3)
    sns.lineplot(data=df[~(df.id.str.startswith("SCHD"))],
                 x="t", y="pending", hue="id", ax=ax)

    ax = plt.subplot(1, 2, 2)
    sns.lineplot(data=df, x="t", y="running", hue="id", ax=ax)

    plt.show()


def rnd(count: int, min_val: int = 1, max_val: int = 20) -> list[int]:
    import random
    return [random.randint(min_val, max_val) for _ in range(count)]


if __name__ == "__main__":
    START_T = 0
    END_T = 500
    EXCS = [
        Executor(id="HPC1", C=20, X=rnd(13), Q=rnd(10)),
        Executor(id="HPC2", C=30, X=rnd(19), Q=rnd(12)),
        Executor(id="HPC3", C=40, X=rnd(23), Q=rnd(10)),
        Executor(id="HPC4", C=40, X=rnd(32), Q=rnd(20)),
        Executor(id="HPC5", C=30, X=rnd(23), Q=rnd(15)),
        Executor(id="HPC6", C=20, X=rnd(17), Q=rnd(16)),
    ]
    SCHD = Scheduler(
        id="SCHD",
        Q=rnd(1000),
        R=EXCS,
        cost_fn=Scheduler.linear_cost,
    )
    SYSTEM: list[Actor] = [SCHD, *EXCS]
    simulate(system=SYSTEM, start_time=START_T, end_time=END_T)
