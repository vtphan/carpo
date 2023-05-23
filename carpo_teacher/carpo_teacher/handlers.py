from distutils.command.config import config
import json
from socket import timeout
import time

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado

import requests
import os
from pathlib import Path
import uuid


def read_config_file():
    """
    reads config.json file
    :return: dict
    """
    config_file = os.path.join(os.getcwd() ,"Exercises",'config.json')
    if os.path.exists(config_file):
        f=open(config_file)
        return json.load(f)
    return {}

def create_initial_files():
    print('======================================')
    print("Check and Create Initial Files:")
    print('======================================')
    current_dir = os.getcwd()
    print(current_dir)
    if "Exercises" not in os.listdir():
        os.makedirs(os.path.join(current_dir,"Exercises"))
   
    # Create config.json file
    config_path = os.path.join(current_dir,"Exercises","config.json")
    if not os.path.isfile(config_path):
        config_data = {}
        config_data['name'] = "John Smith"
        config_data['server'] = "http://delphinus.cs.memphis.edu:XXXX"
        config_data['carpo_version'] = "0.0.9"
        # Write default config
        with open(config_path, "w") as config_file:
            config_file.write(json.dumps(config_data, indent=4))
    
    # Create blank notebook
    notebook_path = os.path.join(current_dir,"Exercises","Readme.ipynb")
    if not os.path.isfile(notebook_path):
        content = {
                        "cells": [],
                        "metadata": {
                            "kernelspec": {
                                "display_name": "Python 3 (ipykernel)",
                                "language": "python",
                                "name": "python3"
                                },
                            "language_info": {
                            "codemirror_mode": {
                                "name": "ipython",
                                "version": 3
                                },
                            "file_extension": ".py",
                            "mimetype": "text/x-python",
                            "name": "python",
                            "nbconvert_exporter": "python",
                            "pygments_lexer": "ipython3",
                            "version": "3.8.10"
                            }
                        },
                        "nbformat": 4,
                        "nbformat_minor": 5
                    }
        
        content["cells"].append({
                                "cell_type": "markdown",
                                "id": str(uuid.uuid4()),
                                "metadata": {},
                                "source": [ "#### To complete carpo installation, do these steps: \n \
1. Edit *config.json* to add your name, and the server address. \n \
2. Click the button **Register**, to register your account. \n" ],
                                "outputs": []
                                })

        with open(notebook_path, "w") as file:
            file.write(json.dumps(content, indent = 4))


class RegistrationHandler(APIHandler):
    def initialize(self,config_files):
        self.config_files = config_files
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()
        if config_data == {}:
            create_initial_files()
            
            self.set_status(500)
            self.finish(json.dumps({'message': "Update your User Name and Server address in Exercises/config.json file and register again."}))
            return
            
        if not {'name','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "Invalid config.json file. Please check your config file."}))
            return
        
        if config_data['name'] == "John Smith":
            self.set_status(500)
            self.finish(json.dumps({'message': "Update your User Name and Server address in Exercises/config.json file and register again."}))
            return
        
        url = config_data['server'] + "/add_teacher"

        body = {}
        body['name'] = config_data['name']

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
       
        try:
            response = requests.post(url, data=json.dumps(body),headers=headers,timeout=5).json()
            print(response)
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        config_data['id'] = response['id']
        config_data['uuid'] = response['uuid']
        # Write id to the json file.
        with open(os.path.join(os.getcwd(),"Exercises",'config.json'), "w") as config_file:
            config_file.write(json.dumps(config_data, indent=4))
        print(response)
        self.finish(response)
        
class SubmissionHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()

        if not {'id', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
            
        id = config_data['id']
        name = config_data['name']
        url = config_data['server'] + "/teachers/submissions" + "?id=" + str(id) + "&name=" + name
        
        try:
            response = requests.get(url,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        # Write response to individual Notebook
        file_paths = self.submission_file(response['data'])
        if file_paths:
            p_list = []
            [ p_list.append(i['Question'] ) for i in file_paths if i['Question'] not in p_list ]
            self.finish(json.dumps({'remaining': response['Remaining'], 'sub_file': ", ".join([ i['Notebook'] for i in file_paths]), 'question': ", ".join(p_list)}))
        else: 
            self.finish(response)
    
    def submission_file(self, data):
        file_paths = []
        for res in data:
            dir_path = os.path.join("Exercises", "problem_{}".format(res['problem_id']))
            file_path = "sub_{:03d}".format(res['id']) + ".ipynb"
            if not os.path.exists(dir_path):
                os.makedirs(dir_path)

            file_paths.append({'Notebook': file_path, 'Question': "{}".format(res['problem_id'])})

            submission_file = os.path.join(dir_path, file_path)
            if not os.path.exists(submission_file):
                content = {
                        "cells": [],
                        "metadata": {
                            "kernelspec": {
                                "display_name": "Python 3 (ipykernel)",
                                "language": "python",
                                "name": "python3"
                                },
                            "language_info": {
                            "codemirror_mode": {
                                "name": "ipython",
                                "version": 3
                                },
                            "file_extension": ".py",
                            "mimetype": "text/x-python",
                            "name": "python",
                            "nbconvert_exporter": "python",
                            "pygments_lexer": "ipython3",
                            "version": "3.8.10"
                            }
                        },
                        "nbformat": 4,
                        "nbformat_minor": 5
                    }

                msg_block = ["## Submission {}\n".format( res['id'])]
                student_msg = ["Student wrote (at {}):  ".format(res['time'])] + [ x.replace("## Message to instructor:", "")+"\n" for x in res['message'].split("\n") ]

                if res['message'].strip(' \t\n\r') != "## Message to instructor:":
                    msg_block = msg_block + student_msg

                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": msg_block
                        })

                code_block = ["#{} {} {}".format(res['student_id'], res['problem_id'], res['id'])]
                content["cells"].append({
                        "cell_type": res['format'],
                        "execution_count": 0,
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": "\n".join(code_block + res['code'].split("\n") ),
                        "outputs": []
                        })

                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": "### Status: New"
                        })
                
                sub_history = ["## Submission History\n"]
                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": sub_history + [ x+"\n" for x in res['info'].split("\n") ] + ["---\n"]
                        })

                # Serializing json 
                json_object = json.dumps(content, indent = 4)

                with open(submission_file, "w") as file:
                    file.write(json_object)
        return file_paths

class GradedSubmissionHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()

        if not {'id', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        url = config_data['server'] + "/teachers/graded_submissions"
        
        try:
            response = requests.get(url,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        # Write response to individual Notebook
        file_paths = self.submission_file(response['data'])
        if file_paths:
            self.finish(json.dumps({'msg':"Graded submissions placed inside "+ ", ".join(file_paths) +"."}))
        else: 
            self.finish(json.dumps({'msg':"New graded submissions not available. Please check again later."}))

    def submission_file(self, data):
        file_paths = []
        for res in data:
            dir_path = os.path.join("Exercises", "problem_{}".format(res['problem_id']),"Graded")
            status = 'c' if res['score'] == 1 else 'i'
            file_path = "{:03d}_{:03d}_{}".format(res['student_id'],res['id'],status) + ".ipynb"
            if not os.path.exists(dir_path):
                os.makedirs(dir_path)
            
            submission_file = os.path.join(dir_path, file_path)
            if not os.path.exists(submission_file):
                file_paths.append(submission_file.replace("Exercises/",""))
                content = {
                        "cells": [],
                        "metadata": {
                            "kernelspec": {
                                "display_name": "Python 3 (ipykernel)",
                                "language": "python",
                                "name": "python3"
                                },
                            "language_info": {
                            "codemirror_mode": {
                                "name": "ipython",
                                "version": 3
                                },
                            "file_extension": ".py",
                            "mimetype": "text/x-python",
                            "name": "python",
                            "nbconvert_exporter": "python",
                            "pygments_lexer": "ipython3",
                            "version": "3.8.10"
                            }
                        },
                        "nbformat": 4,
                        "nbformat_minor": 5
                    }

                msg_block = ["## Submission {}\n".format( res['id'])]
                student_msg = ["Student wrote (at {}):  ".format(res['time'])] + [ x.replace("## Message to instructor:", "")+"\n" for x in res['message'].split("\n") ]

                if res['message'].strip(' \t\n\r') != "## Message to instructor:":
                    msg_block = msg_block + student_msg

                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": msg_block
                        })

                code_block = ["#{} {} {}\n".format(res['student_id'], res['problem_id'], res['id'])]
                content["cells"].append({
                        "cell_type": "code",
                        "execution_count": 0,
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": code_block + [ x+"\n" for x in res['code'].split("\n") ],
                        "outputs": []
                        })

                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in res['comment'].split("\n") ]
                        })

                # Serializing json 
                json_object = json.dumps(content, indent = 4)

                with open(submission_file, "w") as file:
                    file.write(json_object)
        return file_paths

class ProblemHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        input_data['teacher_id'] = config_data['id']
        
        url = config_data['server'] + "/problem"

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers, timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        self.finish(response)

    @tornado.web.authenticated
    def delete(self):
        input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/problem"

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.delete(url, data=json.dumps(input_data),headers=headers, timeout=5)
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return
        self.finish(json.dumps(response.text))
    
class GradeHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/submissions/grade"

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers,timeout=5)

        data = {
            "go-server": response.json()
        }
        self.finish(data)
class FeedbackHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/teachers/feedbacks" 

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        if response['msg'] == "Submission put back into the queue successfully.":
            # Delete the local submission notebook
            notebook_path = os.path.join("Exercises", "problem_{}".format(input_data['problem_id']), "sub_{:03d}".format(input_data['submission_id']) + ".ipynb" )
            if os.path.exists(notebook_path):
                os.remove(notebook_path)


        self.finish(response)

class SolutionHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        url = config_data['server'] + "/solution" + "?problem_id="+str(input_data['problem_id'])

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers, timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        self.finish(response)

class ViewProblemStatusRouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server

    @tornado.web.authenticated
    def get(self):
        # input_data is a dictionary with a key "name"
        # input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id', 'name', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        problems_status_url = config_data['server'] + "/problems/status"

        self.finish({"url":problems_status_url })
    
class GoWebAppRouteHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):
        # input_data is a dictionary with a key "name"
        # input_data = self.get_json_body()
        config_data = read_config_file()

        if not {'id', 'name', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        if not {'app_url'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "app_url not found in config."}))
            return

        web_page_url = config_data['app_url'] +"/#/?token="+ config_data['uuid']

        self.finish({"url":web_page_url })

def setup_handlers(web_app):
    host_pattern = ".*$"

    base_url = web_app.settings["base_url"]
    route_pattern_code = url_path_join(base_url, "carpo-teacher", "submissions")
    handlers = [(route_pattern_code, SubmissionHandler)]
    web_app.add_handlers(host_pattern, handlers)

    route_pattern_problems_status =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "graded_submissions")
    web_app.add_handlers(host_pattern, [(route_pattern_problems_status, GradedSubmissionHandler)])

    route_pattern_problem =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "problem")
    web_app.add_handlers(host_pattern, [(route_pattern_problem, ProblemHandler)])

    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "submissions/grade")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, GradeHandler)])

    route_pattern_feedback =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "submissions/feedbacks")
    web_app.add_handlers(host_pattern, [(route_pattern_feedback, FeedbackHandler)])

    route_pattern_register =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "register")
    web_app.add_handlers(host_pattern, [(route_pattern_register, RegistrationHandler, dict(config_files = create_initial_files()))])

    route_pattern_problems_status =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "view_problem_list")
    web_app.add_handlers(host_pattern, [(route_pattern_problems_status, ViewProblemStatusRouteHandler)])

    route_pattern_problems_status =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "solution")
    web_app.add_handlers(host_pattern, [(route_pattern_problems_status, SolutionHandler)])

    route_pattern_problems_status =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "view_app")
    web_app.add_handlers(host_pattern, [(route_pattern_problems_status, GoWebAppRouteHandler)])

