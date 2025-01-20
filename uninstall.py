import os
import sys
import platform
import shutil
from pathlib import Path

def get_scripts_path():
    if platform.system() == "Windows":
        return "Scripts"
    return "bin"

def remove_venv():
    venv_path = ".venv"
    if os.path.exists(venv_path):
        print("Removing virtual environment...")
        shutil.rmtree(venv_path)

def get_user_scripts_path():
    if platform.system() == "Windows":
        python_home = os.path.dirname(sys.executable)
        return os.path.join(python_home, "Scripts")
    return os.path.expanduser("~/.local/bin")

def remove_command():
    user_scripts = get_user_scripts_path()
    wrapper_path = os.path.join(user_scripts, "ez-utils.cmd" if platform.system() == "Windows" else "ez-utils")
    
    if os.path.exists(wrapper_path):
        print("Removing ez-utils command...")
        os.remove(wrapper_path)

def remove_package_files():
    egg_info = "ez_utils.egg-info"
    if os.path.exists(egg_info):
        print("Removing package metadata...")
        shutil.rmtree(egg_info)

def main():
    print(f"Python version: {platform.python_version()}")
    print(f"Operating system: {platform.system()}")
    
    remove_venv()
    remove_command()
    remove_package_files()
    
    print("\nUninstallation complete!")
    print("The ez-utils package has been removed from your system.")

if __name__ == "__main__":
    main()
