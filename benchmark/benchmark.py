#!/usr/bin/env python3
"""Measure the CLI against real Git and deterministic, offline GitHub responses."""
import argparse
import json
import os
from pathlib import Path
import statistics
import subprocess
import sys
import tempfile
import time


def run(args, cwd, env):
    return subprocess.check_output(args, cwd=cwd, env=env, stderr=subprocess.PIPE, text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', required=True, type=Path)
    parser.add_argument('--compare', type=Path, help='Baseline binary; alternate both binaries on the same fixture')
    parser.add_argument('--branches', type=int, default=30)
    parser.add_argument('--repeat', type=int, default=15)
    parser.add_argument('--scan', choices=['quick', 'deep'], default='quick')
    args = parser.parse_args()
    binary = str(args.binary.resolve())
    if args.branches < 1 or args.repeat < 1:
        parser.error('branches and repeat must be positive')
    with tempfile.TemporaryDirectory(prefix='gh-poi-benchmark-') as directory:
        root = Path(directory)
        repo = root / 'repo'
        repo.mkdir()
        tools = root / 'bin'
        tools.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith(('GIT_', 'GH_'))}
        env.update(GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL=str(root / 'empty-config'),
                   HOME=str(root), GH_HOST='github.com', NO_COLOR='1', TERM='dumb')
        env['PATH'] = str(tools) + os.pathsep + env['PATH']
        git = lambda *parts: run(['git', *parts], repo, env)
        git('init', '-b', 'main')
        git('config', 'user.name', 'Benchmark')
        git('config', 'user.email', 'benchmark@example.invalid')
        git('commit', '--allow-empty', '-m', 'Initial')
        git('remote', 'add', 'origin', 'https://github.com/benchmark/repo.git')
        git('update-ref', 'refs/remotes/origin/main', 'HEAD')
        edges = []
        for i in range(args.branches):
            name = f'topic-{i:03}'
            # Shared history with distinct tips keeps setup small and representative.
            oid = git('commit-tree', 'HEAD^{tree}', '-p', 'main', '-m', name).strip()
            git('update-ref', 'refs/heads/' + name, oid)
            edges.append({'node': {'number': i + 1, 'headRefName': name,
                          'url': f'https://github.com/benchmark/repo/pull/{i + 1}',
                          'state': 'MERGED', 'isDraft': False, 'author': {'login': 'benchmark'},
                          'commits': {'nodes': [{'commit': {'oid': oid}}]}}})
        git('config', 'branch.topic-000.gh-poi-locked', 'true')
        if args.branches > 1:
            git('config', 'branch.topic-001.gh-poi-protected', 'true')
        (root / 'repo.json').write_text(json.dumps({'name': 'repo', 'owner': {'login': 'benchmark'},
                                                  'defaultBranchRef': {'name': 'main'}}))
        (root / 'prs.json').write_text(json.dumps({'data': {'search': {'issueCount': len(edges), 'edges': edges}}}))
        # Only GitHub is stubbed; Git commands, CLI startup, parsing and decisions are real.
        gh = tools / 'gh'
        gh.write_text(f'#!{sys.executable}\n' + '''import json
import os
from pathlib import Path
import re
import sys

root = Path(os.environ['HOME'])
if sys.argv[1:3] == ['repo', 'view']:
    print((root / 'repo.json').read_text())
elif sys.argv[1:3] == ['api', 'graphql']:
    query = next(arg for arg in sys.argv[3:] if arg.startswith('query='))
    hashes = set(re.findall(r'\\b[0-9a-f]{40}\\b', query))
    response = json.loads((root / 'prs.json').read_text())
    search = response['data']['search']
    search['edges'] = [edge for edge in search['edges']
                       if any(node['commit']['oid'] in hashes
                              for node in edge['node']['commits']['nodes'])]
    search['issueCount'] = len(search['edges'])
    print(json.dumps(response))
else:
    sys.exit('unexpected gh arguments')
''')
        gh.chmod(0o755)
        binaries = {'candidate': binary}
        if args.compare:
            binaries['baseline'] = str(args.compare.resolve())
        commands = {name: [path, '--dry-run', '--scan=' + args.scan]
                    for name, path in binaries.items()}
        # Warm each binary before measuring. Correctness belongs to the test suite.
        for command in commands.values():
            run(command, repo, env)
        samples = {name: [] for name in commands}
        for iteration in range(args.repeat):
            order = list(commands)
            if iteration % 2:
                order.reverse()
            for name in order:
                start = time.perf_counter()
                run(commands[name], repo, env)
                samples[name].append((time.perf_counter() - start) * 1000)
        diagnostics = {}
        for name, command in commands.items():
            trace = root / (name + '-trace.json')
            trace_env = dict(env, GIT_TRACE2_EVENT=str(trace))
            run(command, repo, trace_env)
            starts = [json.loads(line) for line in trace.read_text().splitlines()
                      if json.loads(line).get('event') == 'start']
            diagnostics[name] = {'git_commands': len(starts),
                                 'git_config_commands': sum(event['argv'][1] == 'config'
                                                            for event in starts)}
        result = {'branches': args.branches, 'scan': args.scan, 'repeat': args.repeat,
                  'diagnostics': diagnostics,
                  'results': {name: {'median_ms': statistics.median(values),
                                     'min_ms': min(values), 'max_ms': max(values),
                                     'samples_ms': values}
                              for name, values in samples.items()}}
        if args.compare:
            result['improvement_percent'] = 100 * (1 - statistics.median(samples['candidate']) /
                                                   statistics.median(samples['baseline']))
        summary = dict(result)
        summary['results'] = {name: {k: v for k, v in data.items() if k != 'samples_ms'}
                              for name, data in result['results'].items()}
        print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
