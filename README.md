# carpo
### Running carpo_server
1. cd carpo_server
2. go build
3. ./carpo -c config.json
### Requirements to run carpo JupyterLab Extensions

* JupyterLab >= 3.0

Installing carpo_teacher extension

1. pip install --upgrade carpo_teacher

*Update the config.json file inside Carpo directory in the working directory of your Jupyter Lab.
Replace the name and server address.*

### Installing carpo_student extension

1. pip install --upgrade carpo_student

*Update the config.json file inside Carpo directory in the working directory of your Jupyter Lab.
Replace the name and server address provided by your instructor.*

### Setting up development environment for student extension
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
virtualenv dev-student-extension
```
This will create new python virtual environment named `dev-student-extension`

4. Activate the above created virtual environment
```bash
source dev-student-extension/bin/activate
```

5. Install jupyterlab in your environment
```bash
pip install jupyterlab
```

6. Clone the repo into your `Home` directory
```bash
cd ~
git clone https://github.com/vtphan/carpo.git
```

7. Go inside carpo_student directory
```bash
cd carpo/carpo_student
```

8. Build the extension package
```bash
# Install package in development mode
pip install -e .
# Link your development version of the extension with JupyterLab
jupyter labextension develop . --overwrite
# Server extension must be manually installed in develop mode
jupyter server extension enable carpo_student
# Rebuild extension Typescript source after making changes
jlpm build
```

9. Verify the jupyterlab student extension is installed.
```bash
jupyter labextension list
jupyter server extension list
```
Both of these command should show `carpo_student` extension in the list.


10. Launch jupyter lab with student extension
```bash
jupyter lab
```

### Setting up development environment for teacher extension
Follow all the steps mentioned above. You'll have to create a separate virtual environment for teacher extension inside `my-venvs`
```bash
cd ~/my-venvs
virtualenv dev-teacher-extension
source dev-teacher-extension/bin/activate
pip install jupyterlab
cd ~/carpo/carpo_teacher
```
Build the extension package like above and verify the jupyterlab teacher extension is installed.
```bash
jupyter labextension list
jupyter server extension list
```
Both of these command should show `carpo_teacher` extension in the list.

Launch jupyter lab with teacher extension
```bash
jupyter lab
```