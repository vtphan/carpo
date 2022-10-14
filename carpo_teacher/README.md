## Requirements

* JupyterLab >= 3.0

## Installation

1. Open Jupyter Lab
2. Open Terminal
3. Execute this command
```bash
pip install --upgrade carpo_teacher
```
4. Close and restart Jupyter Lab


## Configure to use with carpo server

Teachers/TAs need to edit the **config.json** file in the **Exercises** folder and update the following information:

* Student name
* IP address of the server. The instructor should provide this information.



## Additional installation information

To install specific version (e.g. 0.0.8), execute this command in Jupyter Lab's terminal:

```bash
pip install carpo_teacher==0.0.8
```

To uninstall the exension, execute this command in Jupyter Lab's terminal:

```bash
pip uninstall carpo_teacher
```


## Install the extension with Virtual environment [Optional]
It is recommended that you install the carpo-teacher extension in a separate python virtual environment. This way, you can isolate the carpo modules from your global jupyterlab server.

1. Install python virtual environment
```bash
pip install virtualenv
```
2. Create directory to store your virtual environments for projects
```bash
mkdir ~/my-venvs
cd ~/my-venvs
```
3. Create virtual environment for teacher extension
```bash
virtualenv carpo-teacher-mode
```
This will create new python virtual environment named `carpo-teacher-mode`

4. Activate the above created virtual environment
```bash
source carpo-teacher-mode/bin/activate
```

5. Install jupyterlab and carpo extension in your environment
```bash
pip install --upgrade carpo-teacher
```

6. Now launch  jupterlab
```bash
jupyter lab
```
You will now see all the carpo functionalities when you open Notebook in your jupyter lab server in the browser.
