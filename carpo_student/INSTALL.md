
## Requirements

* JupyterLab >= 3.0

## Installation

1. Open Jupyter Lab
2. Open Terminal
3. Execute this command
```bash
pip install --upgrade carpo_student
```
4. Close and restart Jupyter Lab


## Configure to use with carpo server

Students need to edit the **config.json** file in the **Exercises** folder and update the following information:

* Student name
* IP address of the server. The instructor should provide this information.

Here's an example of the config.json file for a course: [COMP4151](../CodingHub/COMP4151.md)


## Additional installation information

To install specific version (e.g. 0.0.8), execute this command in Jupyter Lab's terminal:

```bash
pip install carpo_student==0.0.8
```

To uninstall the exension, execute this command in Jupyter Lab's terminal:

```bash
pip uninstall carpo_student
```
