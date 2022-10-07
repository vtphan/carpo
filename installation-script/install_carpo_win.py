carpo_student_version = '0.0.8'

import os, stat, subprocess, json

print('1. Installing carpo-student version ' + carpo_student_version)
os.system('pip install carpo-student==' + carpo_student_version + ' --user')

run_file_name = 'RUN_JUPYTER_LAB.bat'
print('2. Installing ' + run_file_name)

conda_info = subprocess.check_output('conda info --json', shell=True)
conda_info_json = json.loads(conda_info)
conda_path = conda_info_json['conda_prefix']

file_content = '''
@echo OFFs

set CONDAPATH={}

set PATH=%PATH%;%CONDAPATH%\Scripts;%CONDAPATH%\Library\\bin

call %CONDAPATH%\Scripts\\activate.bat

rem Launch JupyterLab from the current directory
set current_dir=%cd%

call %CONDAPATH%\Scripts\jupyter-lab.exe  %current_dir%
'''

current_dir = os.getcwd()
output_file = os.path.join(current_dir, run_file_name)
with open(output_file, 'w') as fp:
    fp.write(file_content.format(conda_path))

print('\t', run_file_name,'is created.')
print('\t In the future, to run Jupyter Lab, double click on this file.')
