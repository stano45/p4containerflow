import json
import sys
import subprocess
import tempfile
import os


def update_src_addr(file_path, old_addr, new_addr):
    with tempfile.NamedTemporaryFile(delete=False) as temp_file:
        temp_file_path = temp_file.name

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
        return

    with open(temp_file_path, "w") as file:
        json.dump(data, file, indent=4)

    subprocess.run(
        f"crit encode -i {temp_file_path} -o {file_path}",
        shell=True,
        check=True,
    )

    os.remove(temp_file_path)


if __name__ == "__main__":
    if len(sys.argv) != 4:
        print(
            "Usage: python edit_files_image.py"
            "<file_path> <old_addr> <new_addr>"
        )
        sys.exit(1)

    file_path = sys.argv[1]
    old_addr = sys.argv[2]
    new_addr = sys.argv[3]

    if not os.path.exists(file_path):
        print(f"Error: {file_path} does not exist")
        sys.exit(1)

    if not old_addr or not new_addr:
        print("Error: old_addr and new_addr must not be empty")
        sys.exit(1)

    if old_addr == new_addr:
        print("Error: old_addr and new_addr must be different")
        sys.exit(1)

    update_src_addr(file_path, old_addr, new_addr)
