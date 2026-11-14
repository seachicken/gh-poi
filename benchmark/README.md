# CLI benchmarks

## Run

Build the candidate and run every case:

```sh
go build -o /tmp/gh-poi-after .
python3 benchmark/run.py --binary /tmp/gh-poi-after
```

To compare against a separately built baseline on the same fixture:

```sh
python3 benchmark/run.py --binary /tmp/gh-poi-after --compare /tmp/gh-poi-before
```

For one case:

```sh
python3 benchmark/benchmark.py --binary /tmp/gh-poi-after --compare /tmp/gh-poi-before --scan quick --branches 30
```

Requires Python 3, Git, and a POSIX shell. GitHub responses are provided locally,
so no GitHub login is required. The default is 15 samples per binary;
`--repeat` changes the sample count.

## Cases and measurements

- `cases.json`: all registered cases (Quick/Deep × 3/30/100 topic branches).
- `benchmark.py`: temporary fixture generation and measurement of one case.
- `run.py`: runs every registered case and prints its results.

The fixture uses real Git and fixed offline GitHub responses, with a default
branch, merged topic branches, and locked/legacy-protected branches. User Git
settings are isolated. Fixture files and untimed Git Trace2 diagnostics exist
only in a temporary directory that is removed after the run.

The runner warms both binaries and alternates their execution order. It prints
median/minimum/maximum wall-clock time, Git command counts, and the percentage
change when a baseline is supplied. Setup and diagnostics are excluded from the
measured time.

Measurements cover the complete `--dry-run` CLI process, including startup,
discovery, parsing, deletion decisions and rendering. Actual deletion, remote
pruning and live GitHub latency are outside the scope.
