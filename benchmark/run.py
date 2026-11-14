#!/usr/bin/env python3
"""Run every registered offline CLI benchmark case."""
import argparse
import json
from pathlib import Path
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', required=True, type=Path)
    parser.add_argument('--compare', type=Path)
    parser.add_argument('--repeat', type=int, default=15,
                        help='Samples per binary per case; use 1 only to check the runner')
    args = parser.parse_args()
    if args.repeat < 1:
        parser.error('repeat must be positive')
    directory = Path(__file__).resolve().parent
    cases = json.loads((directory / 'cases.json').read_text())
    for case in cases:
        command = [sys.executable, str(directory / 'benchmark.py'),
                   '--binary', str(args.binary.resolve()),
                   '--scan', case['scan'], '--branches', str(case['branches']),
                   '--repeat', str(args.repeat)]
        if args.compare:
            command.extend(['--compare', str(args.compare.resolve())])
        print(case['name'], flush=True)
        subprocess.run(command, check=True)


if __name__ == '__main__':
    main()
