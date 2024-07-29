#!/usr/bin/env python3

import json
import sys
import subprocess
import tempfile
import os
import tarfile
import shutil


def check_crit_installed():
    try:
        subprocess.run(
            ["crit", "--version"],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
    except (subprocess.CalledProcessError, FileNotFoundError):
        print(
            "Error: 'crit' command not found."
            "Please install CRIU and ensure 'crit' is in your PATH."
        )
        sys.exit(1)
        
def update_src_addr(file_path, old_addr, new_addr):
    try:
        with tempfile.NamedTemporaryFile(delete=False) as temp_file:
            temp_file_path = temp_file.name

        # Decode the image
        subprocess.run(
            f"crit decode -i {file_path} --pretty > {temp_file_path}",
            shell=True,
            check=True,
        )

        with open(temp_file_path, "r") as file:
            data = json.load(file)

        updated = False
        for entry in data.get("entries", []):
            if entry.get("type") == "INETSK":
                src_addrs = entry.get("isk", {}).get("src_addr")
                if old_addr in src_addrs:
                    entry["isk"]["src_addr"] = [new_addr]
                    print(
                        f"Updated src_addr from {old_addr} to "
                        f"{new_addr} in {file_path}"
                    )
                    updated = True

        if not updated:
            print(f"Error: could not find src_addr {old_addr} in {file_path}")
            raise Exception(f"Error: could not find src_addr {old_addr} in {file_path}")

        with open(temp_file_path, "w") as file:
            json.dump(data, file, indent=4)

        # Encode the updated data back into the file
        subprocess.run(
            f"crit encode -i {temp_file_path} -o {file_path}",
            shell=True,
            check=True,
        )
    except Exception as e:
        print(f"An error occurred: {e}")
        # Dump the current data to a file in /tmp for debugging
        error_dump_path = "/tmp/decoded_image.json"
        with open(error_dump_path, "w") as error_file:
            json.dump(data, error_file, indent=4)
        print(f"Decoded image dumped to {error_dump_path} for debugging.")
    finally:
        if os.path.exists(temp_file_path):
            os.remove(temp_file_path)



def process_directory(input_dir, old_addr, new_addr):
    img_file_path = os.path.join(input_dir, "checkpoint", "files.img")
    if os.path.exists(img_file_path):
        update_src_addr(img_file_path, old_addr, new_addr)
    else:
        print(f"Error: {img_file_path} does not exist")


def process_tar(tar_path, old_addr, new_addr):
    with tempfile.TemporaryDirectory() as temp_dir:
        with tarfile.open(tar_path, "r:") as tar:
            tar.extractall(path=temp_dir)

        process_directory(temp_dir, old_addr, new_addr)

        new_tar_path = tar_path + ".new"
        with tarfile.open(new_tar_path, "w:") as tar:
            tar.add(temp_dir, arcname="")

        shutil.move(new_tar_path, tar_path)


if __name__ == "__main__":
    check_crit_installed()

    if len(sys.argv) != 4:
        print(
            "Usage: python edit_files_image.py"
            "<input_dir_or_tar> <old_addr> <new_addr>"
        )
        sys.exit(1)

    input_path = sys.argv[1]
    old_addr = sys.argv[2]
    new_addr = sys.argv[3]

    if not os.path.exists(input_path):
        print(f"Error: {input_path} does not exist")
        sys.exit(1)

    if not old_addr or not new_addr:
        print("Error: old_addr and new_addr must not be empty")
        sys.exit(1)

    if old_addr == new_addr:
        print("Error: old_addr and new_addr must be different")
        sys.exit(1)

    if os.path.isdir(input_path):
        process_directory(input_path, old_addr, new_addr)
    elif input_path.endswith(".tar"):
        process_tar(input_path, old_addr, new_addr)
    else:
        print("Error: input must be a directory or a .tar file")
        sys.exit(1)
