import glob
import subprocess
import os
import shutil


def run_test(*, repo_dir, test_def_dir):
    run_sh = os.path.join(test_def_dir, 'run.sh')
    actual_output = '../actual_output.txt'
    subprocess.run(
        f'{run_sh} > {actual_output}',
        shell=True,
        cwd=repo_dir,
        check=True,
    )

    expected_fs = os.path.join(test_def_dir, 'expected_fs.txt')
    actual_fs = '../actual_fs.txt'
    if os.path.isfile(expected_fs):
        subprocess.run(
            # TODO: Make this command accessible for test creation.
            f'ls -alR --time-style=+ > {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True
        )

        subprocess.run(
            f'diff {expected_fs} {actual_fs}',
            shell=True,
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )

    # TODO: add expected_checksums.txt or expand expected_fs.txt

    expected_output = os.path.join(test_def_dir, 'expected_output.txt')
    if os.path.isfile(expected_output):
        subprocess.run(
            f'diff {expected_output} {actual_output}',
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


if __name__ == '__main__':

    script_dir = os.path.dirname(os.path.realpath(__file__))

    for test_def_dir in glob.iglob(os.path.join(script_dir, 'tests', '*')):
        print(f'{test_def_dir}...', end='')
        try:
            repo_dir = set_up(test_def_dir)
            run_test(
                repo_dir=repo_dir,
                test_def_dir=test_def_dir
            )
            print('OK')
        except subprocess.CalledProcessError as proc:
            print('FAIL')
            print(proc)
            print(proc.output.decode())
