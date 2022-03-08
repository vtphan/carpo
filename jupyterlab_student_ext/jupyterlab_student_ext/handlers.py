import json
from urllib import response

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado
import requests
import os
import uuid


class QuestionRouteHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):
        f=open(os.getcwd()+'/config.json')
        config_data = json.load(f)

        url = config_data['server'] + "/problem?student_id="+str(config_data['id'])
        resp = requests.get(url).json()

        # Create new file:
        filename = 'carpo-problem-'+ str(resp['id']) +'.ipynb'

        # Check if submission file exist or not 
        self.input_file(filename, resp['question'])
        msg = {
            "msg": "{}".format(filename)
        }
        self.finish(json.dumps(msg))

    def input_file(self, filename, question):
        if os.path.exists(filename):
            os.remove(filename)
        
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
                        "source": [ "## Message: \n" ],
                        "outputs": []
                        })
        content["cells"].append({
                        "cell_type": "code",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in question.split("\n") ],
                        "outputs": []
                        })

        json_object = json.dumps(content, indent = 4)
            
        # Writing to Notebook
        with open(filename, "w") as outfile:
            outfile.write(json_object)
        outfile.close()


class FeedbackRouteHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):
        f=open(os.getcwd()+'/config.json')
        config_data = json.load(f)

        # Get latest problem id
        # url = config_data['server'] + "/problem"
        # resp = requests.get(url).json()

        url = config_data['server'] + "/students/get_submission_feedbacks?student_id="+str(config_data['id'])
        print("URL: ", url)
        response = requests.get(url)
        resp = response.json()

        if len(resp['data']) == 0:
            self.finish(json.dumps({
                "msg": "No Feedback available at the moment."
            }))
            return

        # Create new file:
        studnet_name = resp['data'][0]['name']    
        feedback_file = 'feedback_'+ studnet_name + '.ipynb'

        # Create a new feedback file if not exists.
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
        

            # Serializing json 
            json_object = json.dumps(content, indent = 4)
            
            # Writing to sample.json
            with open(feedback_file, "w") as outfile:
                outfile.write(json_object)
            outfile.close()

        # Append new feedbacks to the user
        with open(feedback_file, "r") as file:
            data = json.loads(file.read())
            
        data['cells'] = []
        
        for feedback in resp['data']:
            data["cells"].append({
                    "cell_type": "markdown",
                    "id": str(uuid.uuid4()),
                    "metadata": {},
                    "source": [ x+"\n" for x in feedback['comment'].split("\n") ]
                    })
            data["cells"].append({
                    "cell_type": "code",
                    "id": str(uuid.uuid4()),
                    "metadata": {},
                    "source": [ x+"\n" for x in feedback['code_feedback'].split("\n") ],
                    "outputs": []
                    })

        with open(feedback_file, "w") as file:
                json.dump(data, file)

        msg = {
            "msg": "Latest feedback availabe at {}".format(feedback_file)
        }
        self.finish(json.dumps(msg))

    def input_file(self, filename, question):
        if os.path.exists(filename):
            os.remove(filename)
        
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
                        "cell_type": "code",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in question.split("\n") ],
                        "outputs": []
                        })

        json_object = json.dumps(content, indent = 4)
            
        # Writing to Notebook
        with open(filename, "w") as outfile:
            outfile.write(json_object)
        outfile.close()


class SubmissionRouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server

    @tornado.web.authenticated
    def post(self):
        # input_data is a dictionary with a key "name"
        input_data = self.get_json_body()

        # Read Json file and add infos.
        f=open(os.getcwd()+'/config.json')
        config_data = json.load(f)

        input_data['name'] = config_data['name']
        input_data['course_name'] = config_data['course']
        url = config_data['server'] + "/students/submissions"

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers)

        data = {
            "filepath": "File {}!".format(input_data["name"]),
            "go-server": response.json()
        }
        self.finish(json.dumps(data))


def setup_handlers(web_app):
    host_pattern = ".*$"

    base_url = web_app.settings["base_url"]
    route_pattern = url_path_join(base_url, "jupyterlab-student-ext", "submissions")
    handlers = [(route_pattern, SubmissionRouteHandler)]
    web_app.add_handlers(host_pattern, handlers)


    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "jupyterlab-student-ext", "question")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, QuestionRouteHandler)])


    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "jupyterlab-student-ext", "feedback")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, FeedbackRouteHandler)])



