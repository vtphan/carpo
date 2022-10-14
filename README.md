Carpo is a Jupyter-lab extension that supports in-class coding. It's designed to facilitate active learning in Data Science and Machine Learning courses.

To use carpo, instructors, teaching assistants and students must first install JupyterLab (version 3.0 or later) on their computers.

## Overview of system architecture and setup
There are two ways of using carpo for in-class coding.

1. Run carpo_server on a dedicated server. Teachers, TAs, students install appropriate Jupyter extensions on their computers.

2. Run carpo_server on a dedicated server. Students can log into a coding hub to do in-class exercises. Teachers and TA install appropriate Jupyter extensions on their computers.

## Installation instructions

## Setting up carpo_server

The server is written in Go.  You can either download the source and compile it on your Linux 
server.  Or you can download the binary executable.

Installation instructions can be found [here](carpo_server/README.md).


## Installation of Jupyter Lab extensions

### For students
Students can use the CodingHub without having to install the JupyterLab extension on their computers.  See [COMP4151](CodingHub/COMP4151.md), for example, on how to login to the CodingHub.

Alternately, students can install the extension on the own computer.
* [Student installation instructions](carpo_student/INSTALL.md)

### For teachers and TAs
Teachers and teaching assistants need to install the following extension on their computers.
* [Teacher installation instructions](carpo_teacher/README.md)
