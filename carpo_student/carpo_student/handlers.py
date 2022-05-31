import json
from urllib import response

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado
import requests
import os
import uuid


def read_config_file():
    """
    reads config.json file
    :return: dict
    """
    config_file = os.path.join(os.getcwd() ,"Carpo",'config.json')
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
    if "Carpo" not in os.listdir():
        os.makedirs(os.path.join(current_dir,"Carpo"))
   
    # Create config.json file
    config_path = os.path.join(current_dir,"Carpo","config.json")
    if not os.path.isfile(config_path):
        config_data = {}
        config_data['name'] = "John Smith"
        config_data['server'] = "http://delphinus.cs.memphis.edu:XXXX"
        config_data['carpo_version'] = "0.0.1"
        # Write default config
        with open(config_path, "w") as config_file:
            config_file.write(json.dumps(config_data, indent=4))
    
    # Create blank notebook
    notebook_path = os.path.join(current_dir,"Carpo","Readme.ipynb")
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
2. Click the button **Register Carpo**, to register your account. \n" ],
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
            self.finish(json.dumps({'message': "Update your User Name and Server address in Carpo/config.json file and register again."}))
            return 
            
        if not {'name','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "Invalid config.json file. Please check your config file."}))
            return
        
        if config_data['name'] == "John Smith":
            self.set_status(500)
            self.finish(json.dumps({'message': "Update your User Name and Server address in Carpo/config.json file and register again."}))
            return

        if not {'name','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "Invalid config.json file. Please check your config file."}))
            return
        
        url = config_data['server'] + "/add_student"

        body = {}
        body['name'] = config_data['name']

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        
        try:
            response = requests.post(url, data=json.dumps(body),headers=headers,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        config_data['id'] = response['id']
        # Write id to the json file.
        with open(os.path.join(os.getcwd(),"Carpo",'config.json'), "w") as config_file:
            config_file.write(json.dumps(config_data, indent=4))

        self.finish(response)

class QuestionRouteHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()

        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        url = config_data['server'] + "/problem?student_id="+str(config_data['id'])
        try:
            resp = requests.get(url,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        # Write questions to individual Notebook
        file_paths = self.question_file(resp['data'])
        msg = ""
        print(file_paths)
        if file_paths['new_download']:
            msg = "New Problem(s) downloaded and placed in notebook(s) " + ', '.join(file_paths['new_download']) + '.'

        if file_paths['already_downloaded']:
            msg += "\nProblem(s) already downloaded and placed in notebook(s) " + ', '.join(file_paths['already_downloaded']) + '.'

        if len(resp['data']) == 0: 
            msg = "You have got 0 new problems. Please check again later."
    
        self.finish(json.dumps({'msg': msg}))

    def question_file(self, data):
        file_paths = {}
        file_paths['new_download'] = []
        file_paths['already_downloaded'] = []
        for res in data:
            file_path = os.path.join(os.getcwd(),"Carpo","p{:03d}".format( res['id']) + ".ipynb")

            if not os.path.exists(file_path):
                file_paths['new_download'].append("p{:03d}".format( res['id']) + ".ipynb")
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
                                "source": [ "## Message to instructor: \n" ],
                                "outputs": []
                                })
                problem_block = ["## PID {}\n## Expires at {}\n".format(res['id'], res['lifetime'])]
                content["cells"].append({
                                "cell_type": res['format'],
                                "execution_count": 0,
                                "id": str(uuid.uuid4()),
                                "metadata": {},
                                "source": problem_block + [ x+"\n" for x in res['question'].split("\n") ],
                                "outputs": []
                                })

                # Serializing json 
                json_object = json.dumps(content, indent = 4)

                with open(file_path, "w") as file:
                    file.write(json_object)
            else:
                file_paths['already_downloaded'].append("p{:03d}".format( res['id']) + ".ipynb")


        return file_paths

class FeedbackRouteHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):
        config_data = read_config_file()

        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        url = config_data['server'] + "/students/get_submission_feedbacks?student_id="+str(config_data['id'])
        
        try:
            response = requests.get(url,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        if len(response['data']) == 0:
            self.finish(json.dumps({
                "msg": "No Feedback available at the moment. Please check again later."
            }))
            return
        
        # Write feedbacks to individual Notebook
        file_paths = self.feedback_file(response['data'])
        if file_paths:
            msg = "Feedback(s) placed in " + ','.join(file_paths) + '.'
        
        self.finish(json.dumps({'msg':msg}))

    def feedback_file(self, data):
        file_paths = []
        for res in data:
            dir_path = os.path.join("Carpo","Feedback")
            file_path = "p{:03d}_{:03d}".format(res['problem_id'],res['id']) + ".ipynb"
            file_paths.append("Feedback/" + file_path)

            if not os.path.exists(dir_path):
                os.makedirs(dir_path)

            feedback_file = os.path.join(dir_path, file_path)
            if os.path.exists(feedback_file):
                os.remove(feedback_file)

            if not os.path.exists(feedback_file):
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
                        "source": [ x+"\n" for x in res['message'].split("\n") ]
                        })


                content["cells"].append({
                        "cell_type": "code",
                        "execution_count": 0,
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in res['code_feedback'].split("\n") ],
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

                with open(feedback_file, "w") as file:
                    file.write(json_object)
        return file_paths
class SubmissionRouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server

    @tornado.web.authenticated
    def post(self):
        # input_data is a dictionary with a key "name"
        input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id', 'name', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['name'] = config_data['name']
        url = config_data['server'] + "/students/submissions"


        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        self.finish(response)

class ViewStatusRouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server

    @tornado.web.authenticated
    def get(self):
        # input_data is a dictionary with a key "name"
        input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id', 'name', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        id = config_data['id']
        name = config_data['name']
        student_status_url = config_data['server'] + "/students/status" + "?student_id=" + str(id) + "&student_name=" + name

        self.finish({"url":student_status_url })

class ViewProblemStatusRouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server

    @tornado.web.authenticated
    def get(self):
        # input_data is a dictionary with a key "name"
        input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id', 'name', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        problems_status_url = config_data['server'] + "/problems/status"

        self.finish({"url":problems_status_url })
 

def setup_handlers(web_app):
    host_pattern = ".*$"

    base_url = web_app.settings["base_url"]
    route_pattern = url_path_join(base_url, "carpo-student", "submissions")
    handlers = [(route_pattern, SubmissionRouteHandler)]
    web_app.add_handlers(host_pattern, handlers)

    route_pattern_register =  url_path_join(web_app.settings['base_url'], "carpo-student", "register")
    web_app.add_handlers(host_pattern, [(route_pattern_register, RegistrationHandler, dict(config_files = create_initial_files()))])

    route_pattern_question =  url_path_join(web_app.settings['base_url'], "carpo-student", "question")
    web_app.add_handlers(host_pattern, [(route_pattern_question, QuestionRouteHandler)])


    route_pattern_feedback =  url_path_join(web_app.settings['base_url'], "carpo-student", "feedback")
    web_app.add_handlers(host_pattern, [(route_pattern_feedback, FeedbackRouteHandler)])

    route_pattern_view_status =  url_path_join(web_app.settings['base_url'], "carpo-student", "view_student_status")
    web_app.add_handlers(host_pattern, [(route_pattern_view_status, ViewStatusRouteHandler)])

    route_pattern_problems_status =  url_path_join(web_app.settings['base_url'], "carpo-student", "view_problem_list")
    web_app.add_handlers(host_pattern, [(route_pattern_problems_status, ViewProblemStatusRouteHandler)])




