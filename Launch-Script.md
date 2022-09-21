## Windows Launch Script With Anaconda
#### Make sure anaconda is installed & find where it is installed.
1. Go to Start.
2. Type: anaconda. Click on Anaconda Prompt (anaconda3). This will open a command prompt.
3. Type: where anaconda. This will show where anaconda is installed on your computer. Copy the path:
> C:\Users\thomas\anaconda3\Scripts\anaconda.exe

#### Organize directory and install extension
1. Create a new directory anywhere you want and give it a name: COMPXXX. It is better if your put all notebooks for the course in single directory in Desktop or Documents or any drive.

2. Create a New File using Notepad from start.
3. Copy paste the following code into the notepad. Use the path from above `step-3` for __set CONDAPATH=__ but remove: `"\Scripts\anaconda.exe"`
```batch
@echo OFF

set CONDAPATH=C:\Users\thomas\anaconda3

set PATH=%PATH%;%CONDAPATH%\Scripts;%CONDAPATH%\Library\bin

call %CONDAPATH%\Scripts\activate.bat
call pip install carpo-student==0.0.8

rem Launch JupyterLab from the current directory
set current_dir=%cd%

call %CONDAPATH%\Scripts\jupyter-lab.exe  %current_dir%
```
4. Use `Save As` to save the file inside the directory created in `Step 1`. Give it a name: `launch-jupyter.bat`. This file should be executable when saved as .bat
5. Double click on the `launch-jupyter.bat`. It should launch Jupyter Lab from your directory. The `carpo-student` extension should also be available.

## MacOS Launch Script
#### Make sure jupyter lab is installed.
1. Create a new directory anywhere you want and give it a name: COMPXXX. It is better if your put all notebooks for the course in single directory in Desktop or Documents or any drive.

2. Create a New File and save it is `launch-jupyter.sh`
```bash
cd `pwd`;
pip install carpo-student==0.0.8
jupyter lab
```
