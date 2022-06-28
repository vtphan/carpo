
### Requirements to run carpo JupyterLab Extensions

* JupyterLab >= 3.0

### Installing **carpo_student** extension

```bash
pip install --upgrade jupyterlab carpo_student
```

*Update the config.json file inside Carpo directory in the working directory of your Jupyter Lab.
Replace the name and server address provided by your instructor.*

To install specific version use:

>*pip install carpo_student==0.0.4*

### Installation with Virtual Environment
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
virtualenv carpo-mode
```
This will create new python virtual environment named `carpo-mode`

4. Activate the above created virtual environment
```bash
source carpo-mode/bin/activate
```

5. Install jupyterlab and carpo extension in your environment
```bash
pip install --upgrade jupyterlab carpo-student
```
6. Now launch  jupterlab
```bash
jupyter lab
```
You will now see all the carpo functionalities when you open Notebook in your jupyter lab server in the browser.

### Uninstalling **carpo_student** extension
 ```bash
pip uninstall carpo-student
```
### Installing **carpo_teacher** extension [Only for teacher]

```bash
pip install --upgrade jupyterlab carpo_teacher

```

*Update the config.json file inside Carpo directory in the working directory of your Jupyter Lab.
Replace the name and server address.*


