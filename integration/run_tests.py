import glob
import subprocess
import os
import shutil
import sys


DIFF_CMD = 'diff -b --color=always'


def run_test(*, repo_dir, test_def_dir):
    run_sh = os.path.join(test_def_dir, 'run.sh')
    actual_output = '../actual_output.txt'
    subprocess.run(
        f'{run_sh} > {actual_output}',
        shell=True,
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    expected_fs = os.path.join(test_def_dir, 'expected_fs.txt')
    actual_fs = '../actual_fs.txt'
    if os.path.isfile(expected_fs):
        subprocess.run(
            # TODO: Make this command accessible for test creation.
            f'find . -printf "%M %u %g %7s %p   %l\n" > {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True
        )

        subprocess.run(
            f'{DIFF_CMD} {expected_fs} {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )

    expected_duclock = os.path.join(test_def_dir, 'expected_duclock.yaml')
    actual_duclock = './Ducfile.lock'
    if os.path.isfile(expected_duclock):
        subprocess.run(
            f'{DIFF_CMD} {expected_duclock} {actual_duclock}',
            shell=True,
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )

    # TODO: add expected_checksums.txt or expand expected_fs.txt

    expected_output = os.path.join(test_def_dir, 'expected_output.txt')
    if os.path.isfile(expected_output):
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


def _normalize_paths(paths):
    return sorted(os.path.normpath(path) for path in paths)


def run_tests(test_def_dir):
    output_width = 60
    sub_dirs = _normalize_paths(glob.glob(os.path.join(test_def_dir, '*', '')))
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
    return True


if __name__ == '__main__':
    if len(sys.argv) == 1:
        script_dir = os.path.dirname(os.path.realpath(__file__))
        test_dirs = glob.iglob(os.path.join(script_dir, 'tests', '*'))
    else:
        # os.path.normpath, among other things, removes trailing slashes, which
        # is key when using os.path.join.
        test_dirs = sys.argv[1:]

    all_successful = True
    for test_def_dir in _normalize_paths(test_dirs):
        repo_dir = set_up(test_def_dir)
        all_successful &= run_tests(test_def_dir)

    sys.exit(0 if all_successful else 1)
