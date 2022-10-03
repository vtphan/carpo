
## Requirements

* JupyterLab >= 3.0

The following `pip` command(s) can be run from your terminal or from JupyterLab terminal.
## Install

To install the extension, execute:

```bash
pip install carpo_student
```
## Update
To update the extension to latest version, run:
```bash
pip install --upgrade carpo_student
```
## Install specific version
```bash
pip install carpo_student==X.X.X
```

## Uninstall
To remove the extension, execute:

```bash
pip uninstall carpo_student
```

## Configure to use with carpo server

*Update the config.json file inside Carpo directory in the working directory of your Jupyter Lab.
Replace the name and server address.*

## Install the extension with Virtual environment [Optional]
It is recommended that you install the carpo-student extension in a separate python virtual environment. This way, you can isolate the carpo modules from your global jupyterlab server.

1. Install python virtual environment
```bash
pip install virtualenv
```
2. Create directory to store your virtual environments for projects
```bash
mkdir ~/my-venvs
cd ~/my-venvs
```
3. Create virtual environment for student extension
```bash
virtualenv carpo-student-mode
```
This will create new python virtual environment named `carpo-student-mode`

4. Activate the above created virtual environment
```bash
source carpo-student-mode/bin/activate
```

5. Install jupyterlab and carpo extension in your environment
```bash
pip install --upgrade carpo-student
```

6. Now launch  jupterlab
```bash
jupyter lab
```
You will now see all the carpo functionalities when you open Notebook in your jupyter lab server in the browser.
