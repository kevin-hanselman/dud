import os
import time
import sys
import subprocess
import glob
import argparse
import shutil

from run_tests import normalize_paths


def run_benchmark(bench_def_dir):
    # TODO: add support for setup_<cmd>.sh to isolate the commands to benchmark
    # (e.g. commit) from boilerplate.
    for cmd in ['duc', 'dvc']:
        run_sh = os.path.join(bench_def_dir, f'run_{cmd}.sh')
        if os.path.isfile(run_sh):
            scratch_dir = setup(bench_def_dir)
            print(f'Running {run_sh!r}...')
            sys.stdout.flush()
            start = time.time()
            try:
                subprocess.run(
                    run_sh,
                    shell=True,
                    cwd=scratch_dir,
                    check=True,
                    capture_output=True,
                )
                elapsed = time.time() - start
                print(f'Elapsed time: {elapsed:g}')
            except subprocess.CalledProcessError as proc:
                print(proc)
                if proc.stdout:
                    print('-STDOUT-')
                    print(proc.stdout.decode())
                if proc.stderr:
                    print('-STDERR-')
                    print(proc.stderr.decode())


def _copy_data(bench_def_dir, scratch_dir):
    data_src = os.path.join(bench_def_dir, "data")
    data_dst = os.path.join(scratch_dir, "data")
    print(f'Copying {data_src!r} to {data_dst!r}...')
    sys.stdout.flush()
    shutil.copytree(data_src, data_dst)
    subprocess.run('sync', check=True)


def setup(bench_def_dir):
    scratch_dir = os.path.realpath(os.path.join('.', os.path.basename(bench_def_dir)))
    if os.path.exists(scratch_dir):
        print(f'Deleting {scratch_dir!r}...')
        shutil.rmtree(scratch_dir)
    os.makedirs(scratch_dir)
    _copy_data(bench_def_dir, scratch_dir)
    return scratch_dir


def parse_cli_args():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument(
        'benchmark_dirs',
        nargs='*',
        help='benchmark directories to run (if empty, run all)'
    )
    return parser.parse_args()


if __name__ == '__main__':
    cli_args = parse_cli_args()
    if not cli_args.benchmark_dirs:
        script_dir = os.path.dirname(os.path.realpath(__file__))
        cli_args.benchmark_dirs = glob.iglob(os.path.join(script_dir, 'benchmarks', '*'))

    for bench_def_dir in normalize_paths(cli_args.benchmark_dirs):
        print(f'Running {bench_def_dir!r}...')
        run_benchmark(bench_def_dir)
