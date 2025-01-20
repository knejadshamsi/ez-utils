import os
import sys
import platform
import subprocess
import site
import shutil
from pathlib import Path

def get_python_command():
    return sys.executable

def get_scripts_path():
    if platform.system() == "Windows":
        return "Scripts"
    return "bin"

def create_venv(venv_path):
    python = get_python_command()
    subprocess.run([python, "-m", "venv", venv_path], check=True)

def get_venv_python(venv_path):
    scripts = get_scripts_path()
    if platform.system() == "Windows":
        return os.path.join(venv_path, scripts, "python.exe")
    return os.path.join(venv_path, scripts, "python")

def install_package(venv_python):
    subprocess.run([venv_python, "-m", "pip", "install", "-e", "."], check=True)

def get_user_scripts_path():
    if platform.system() == "Windows":
        python_home = os.path.dirname(sys.executable)
        return os.path.join(python_home, "Scripts")
    return os.path.expanduser("~/.local/bin")

def setup_command(venv_path, user_scripts):
    scripts = get_scripts_path()
    venv_scripts = os.path.join(venv_path, scripts)
    os.makedirs(user_scripts, exist_ok=True)
    
    if platform.system() == "Windows":
        wrapper_path = os.path.join(user_scripts, "ez-utils.cmd")
        exe_path = os.path.abspath(os.path.join(venv_scripts, "ez-utils.exe"))
        with open(wrapper_path, "w") as f:
            f.write("@echo off\n")
            f.write(f'if not exist "{exe_path}" (\n')
            f.write('    echo Error: ez-utils executable not found.\n')
            f.write('    echo Please reinstall the package.\n')
            f.write('    exit /b 1\n')
            f.write(')\n')
            f.write(f'"{exe_path}" %*')
    else:
        wrapper_path = os.path.join(user_scripts, "ez-utils")
        with open(wrapper_path, "w") as f:
            f.write("#!/bin/bash\n")
            f.write(f'exec "{os.path.join(venv_scripts, "ez-utils")}" "$@"')
        os.chmod(wrapper_path, 0o755)

def main():
    print(f"Python version: {platform.python_version()}")
    print(f"Operating system: {platform.system()}")

    venv_path = ".venv"
    user_scripts = get_user_scripts_path()
    
    if os.path.exists(venv_path):
        print("Removing existing virtual environment...")
        shutil.rmtree(venv_path)
    
    print("Creating virtual environment...")
    create_venv(venv_path)
    
    venv_python = get_venv_python(venv_path)
    
    print("Installing package...")
    install_package(venv_python)
    
    print("Setting up ez-utils command...")
    setup_command(venv_path, user_scripts)
    
    print("\nInstallation complete!")
    print("You can now use 'ez-utils' command from anywhere.")
    if platform.system() == "Windows":
        print("\nNote: If 'ez-utils' is not recognized, ensure Python Scripts directory is in your PATH:")
        print(f"1. Open System Properties > Advanced > Environment Variables")
        print(f"2. Add '{user_scripts}' to your User PATH")
        print(f"3. Restart your terminal")
    else:
        print("\nNote: You may need to restart your terminal or run:")
        print(f"export PATH=\"$HOME/.local/bin:$PATH\"")

if __name__ == "__main__":
    main()
