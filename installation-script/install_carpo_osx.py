carpo_student_version = '0.0.8'

import os, stat
print('1. Installing carpo-student version ' + carpo_student_version)
os.system('pip install carpo-student==' + carpo_student_version)

run_file_name = 'RUN_JUPYTER_LAB.command'
print('2. Installing ' + run_file_name)

file_content = '''
cd "{}";
jupyter lab
'''
root_dir = os.getcwd()
output_file = os.path.join(root_dir, run_file_name)
with open(output_file, 'w') as fp:
    fp.write(file_content.format(root_dir))
st = os.stat(output_file)
os.chmod(output_file, stat.S_IWRITE | stat.S_IREAD | stat.S_IEXEC)

print('\t', run_file_name,'is created.')
print('\t In the future, to run Jupyter Lab, double click on this file.')