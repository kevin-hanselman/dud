import argparse
import glob
import subprocess
import os
import shutil
import sys
# TODO: consider switching to pathlib


DIFF_CMD = 'git diff --no-index -b --word-diff=color --color=always'
# Replace the username with a generic "user" for compatibility across systems.
# Strangely, --sort=version is needed to exactly match the output of the Github
# Action runner. The Github Action runner's output seems unfazed by a manual
# '| sort -k4'. It may have something to do with hidden files being sorted
# differently on different systems?
FS_CMD = 'tree -afisup --sort=version --noreport | sed "s/$(whoami)/user/"'


def run_test(*, repo_dir, test_def_dir, pin=False):
    run_sh = os.path.join(test_def_dir, 'run.sh')
    expected_output = os.path.join(test_def_dir, 'expected_output.txt')
    actual_output = '../actual_output.txt'

    subprocess.run(
        f'{run_sh} > {actual_output}',
        shell=True,
        cwd=repo_dir,
        check=True,
        capture_output=True,
        timeout=5,
    )

    expected_fs = os.path.join(test_def_dir, 'expected_fs.txt')
    actual_fs = '../actual_fs.txt'
    if os.path.isfile(expected_fs):
        subprocess.run(
            f'{FS_CMD} > {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True
        )
        if pin:
            actual_fs_abs = os.path.realpath(os.path.join(repo_dir, actual_fs))
            shutil.copyfile(actual_fs_abs, expected_fs)

        subprocess.run(
            f'{DIFF_CMD} {expected_fs} {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )

    if os.path.isfile(expected_output):
        if pin:
            actual_output_abs = os.path.realpath(os.path.join(repo_dir, actual_output))
            shutil.copyfile(actual_output_abs, expected_output)
        subprocess.run(
            f'{DIFF_CMD} {expected_output} {actual_output}',
            shell=True,
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )


def set_up(test_def_dir):
    scratch_dir = os.path.realpath(os.path.join('.', os.path.basename(test_def_dir)))
    repo_dir = os.path.join(scratch_dir, 'repo')
    os.makedirs(repo_dir)
    return repo_dir


def normalize_paths(paths):
    # os.path.normpath, among other things, removes trailing slashes, which
    # is key when using os.path.join.
    return sorted(os.path.normpath(path) for path in paths)


def run_tests(test_def_dir, pin=False):
    output_width = 60
    sub_dirs = normalize_paths(glob.glob(os.path.join(test_def_dir, '*', '')))
    has_sub_dirs = len(sub_dirs) > 0

    test_name = os.path.basename(test_def_dir)
    if has_sub_dirs:
        print(test_name)
    else:
        print(f'{test_name:.<{output_width}}', end='')
        sub_dirs = [test_def_dir]

    for sub_dir in sub_dirs:
        if has_sub_dirs:
            print(f'  {os.path.basename(sub_dir):.<{output_width - 2}}', end='')
        try:
            run_test(
                repo_dir=repo_dir,
                test_def_dir=sub_dir,
                pin=pin,
            )
            print('OK')
        except subprocess.CalledProcessError as proc:
            print('FAIL\n')
            print(proc)
            if proc.stdout:
                print('-STDOUT-')
                print(proc.stdout.decode())
            if proc.stderr:
                print('-STDERR-')
                print(proc.stderr.decode())
            return False  # stop running sub-tests on a failure
        except Exception as ex:
            print('ERR')
            print(ex)
            return False
    return True


def parse_cli_args():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument(
        'test_dirs',
        nargs='*',
        help='top-level test directories to run (if empty, run all)'
    )
    parser.add_argument(
        '--pin',
        action='store_true',
        help='if set, overwrite any existing targets with their current values'
    )
    return parser.parse_args()


if __name__ == '__main__':
    cli_args = parse_cli_args()
    if not cli_args.test_dirs:
        script_dir = os.path.dirname(os.path.realpath(__file__))
        cli_args.test_dirs = glob.iglob(os.path.join(script_dir, 'tests', '*'))

    # Using a fixed path is important for consistent filesystem listing output
    # on different systems. For the same reason the umask must be fixed to
    # ensure new files and directories show the same permissions.
    working_dir = '/tmp/dud_integration_tests'
    os.makedirs(working_dir, exist_ok=True)
    os.chdir(working_dir)
    # Default group and others to read-only. Ubuntu's default is apparently
    # 0o0002 (see: https://askubuntu.com/a/44548).
    os.umask(0o0022)

    all_successful = True
    for test_def_dir in normalize_paths(cli_args.test_dirs):
        repo_dir = set_up(test_def_dir)
        all_successful &= run_tests(test_def_dir, pin=cli_args.pin)

    sys.exit(0 if all_successful else 1)
