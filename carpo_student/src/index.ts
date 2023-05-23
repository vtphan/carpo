import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import {
  INotebookTracker,
  // NotebookActions,
  NotebookPanel,
  INotebookModel,

} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { PanelLayout } from '@lumino/widgets';

import {
  CellCheckButton
} from './widget'

import { CellInfo } from './model'

import { ISettingRegistry } from '@jupyterlab/settingregistry';

import { requestAPI } from './handler';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';

import { ToolbarButton,Dialog, showDialog,showErrorMessage } from '@jupyterlab/apputils';

import { DocumentRegistry } from '@jupyterlab/docregistry';

import { ShareCodeButton } from './share-code'
import { GetSolutionButton } from './get-solutions'


/**
 * Initialization data for the carpo-student extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'carpo-student:plugin',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [ISettingRegistry],
  activate: (
      app: JupyterFrontEnd, 
      nbTrack: INotebookTracker,
      settingRegistry: ISettingRegistry | null
    ) => {
    console.log('JupyterLab extension carpo-student is activated!');
    // var interval => (); 
    var counter: number = 0;
    nbTrack.currentChanged.connect(() => {
      console.log("my counter: ", counter);
      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;
      const filename = notebookPanel.context.path

      // Disable Code Share functionality if inside Feedback directory
      if (filename.includes("Feedback")){
        return 
      }

      // Disable if not inside Exercises directory
      if (!filename.includes("Exercises")) {
        return
      }

      notebookPanel.context.ready.then(async () => {

        let currentCell: Cell = null;
        let currentCellCheckButton: CellCheckButton = null;

        nbTrack.activeCellChanged.connect(() => {

          var question:string

          if (currentCell) {
            notebook.widgets.map((c: Cell) => {
              if (c.model.type == 'code' || c.model.type == 'markdown' ) {
                const currentLayout = c.layout as PanelLayout;
                currentLayout.widgets.map(w => {
                  if (w === currentCellCheckButton) {
                    currentLayout.removeWidget(w)
                  }
                })
              }
            });
          }

          const cell: Cell = notebook.activeCell;
          const activeIndex = notebook.activeCellIndex

          var info : CellInfo = {
            problem_id: parseInt((filename.split("/").pop()).replace("ex","").replace(".ipynb",""))
          };

          // Get the message block referencing the active cell.
          notebook.widgets.map((c,index) =>{
            if (c.model.value.text.startsWith("## Message to instructor:")){
              info.message = c.model.value.text
            }
            if (index == activeIndex) {
              question = c.model.value.text
              if (question.includes("## PID ")){

                const newCheckButton: CellCheckButton = new CellCheckButton(cell,info);

                (cell.layout as PanelLayout).addWidget(newCheckButton);
                currentCellCheckButton = newCheckButton;

                // Send code snapshot to the server:
                if (counter == 0 ){
                  setInterval(function () {
                    let postBody = {
                      "message": "",
                      "code": c.model.value.text,
                      "problem_id":info.problem_id,
                      "snapshot": 1
                      }
                      requestAPI<any>('submissions',{
                          method: 'POST',
                          body: JSON.stringify(postBody)
                      })
                      .then(data => {
                          console.log("Snapshot sent.", data)
                        });
                  }, 20000);
                  counter ++;
                }
              }
            }
          })
       

          // const newCheckButton: CellCheckButton = new CellCheckButton(cell,info);
          
          // if (question.includes("## PID ")){
          //   (cell.layout as PanelLayout).addWidget(newCheckButton);
          //   currentCellCheckButton = newCheckButton;
          // }

          // Set the current cell and button for future
          // reference
          currentCell = cell;

        });

      });
    });


    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new RegisterButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    app.docRegistry.addWidgetExtension('Notebook', new ShareCodeButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetFeedbackButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetSolutionButton());
    app.docRegistry.addWidgetExtension('Notebook', new ViewSubmissionStatusButton());
    // app.docRegistry.addWidgetExtension('Notebook', new viewProblemStatusExtension());


  }
};
export class RegisterButton
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  /**
   * Create a new extension for the notebook panel widget.
   *
   * @param panel Notebook panel
   * @param context Notebook context
   * @returns Disposable on the added button
   */
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const register = () => {
      // NotebookActions.clearAllOutputs(panel.content);

      // const notebook = panel.content;

      requestAPI<any>('register',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title:'',
            body:  "Student "+ data.name + " has been registered.",
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Registration Error', reason);
          console.error(
            `Failed to register user as Student.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'register-button',
      label: 'Register',
      onClick: register,
      tooltip: 'Register as a Student',
    });

    panel.toolbar.insertItem(10, 'register', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class GetQuestionButton
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  /**
   * Create a new extension for the notebook panel widget.
   *
   * @param panel Notebook panel
   * @param context Notebook context
   * @returns Disposable on the added button
   */
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const getQuestion = () => {
      // NotebookActions.clearAllOutputs(panel.content);

      // const notebook = panel.content;

      requestAPI<any>('question',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title:'',
            body:  data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Get Problem Error', reason);
          console.error(
            `Failed to get active questions.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-question-button',
      label: 'GetProblem',
      onClick: getQuestion,
      tooltip: 'Get Latest Problem From Server',
    });

    panel.toolbar.insertItem(11, 'getQuestion', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}
export class GetFeedbackButton
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  /**
   * Create a new extension for the notebook panel widget.
   *
   * @param panel Notebook panel
   * @param context Notebook context
   * @returns Disposable on the added button
   */
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const getFeedback = () => {

      requestAPI<any>('feedback',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);
          showDialog({
            title:'',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          }).then( result => {
            if (result.button.accept && data['hard-reload'] == 1 ) {
                window.location.reload();
            }
          })
        })
        .catch(reason => {
          showErrorMessage('Get Feedback Error', reason);
          console.error(
            `Failed to fetch recent feedbacks.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-feedback-button',
      label: 'GetFeedback',
      onClick: getFeedback,
      tooltip: 'Get Feedback to your Submission',
    });

    panel.toolbar.insertItem(13, 'getFeedback', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class ViewSubmissionStatusButton
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  /**
   * Create a new extension for the notebook panel widget.
   *
   * @param panel Notebook panel
   * @param context Notebook context
   * @returns Disposable on the added button
   */
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const viewStatus = () => {

      requestAPI<any>('view_student_status',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);
          window.open(
            data.url, "_blank");
        })
        .catch(reason => {
          showErrorMessage('View Status Error', reason);
          console.error(
            `Failed to view student submission status.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-status-button',
      label: 'Status',
      onClick: viewStatus,
      tooltip: 'View your submissions status',
    });

    panel.toolbar.insertItem(14, 'viewStatus', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

// Currently disabled
export class viewProblemStatusExtension
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  /**
   * Create a new extension for the notebook panel widget.
   *
   * @param panel Notebook panel
   * @param context Notebook context
   * @returns Disposable on the added button
   */
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const viewProblemStatus = () => {

      requestAPI<any>('view_problem_list',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);
          window.open(
            data.url, "_blank");
        })
        .catch(reason => {
          showErrorMessage('View Problem Status Error', reason);
          console.error(
            `Failed to view problem status.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-status-button',
      label: 'Problems',
      onClick: viewProblemStatus,
      tooltip: 'View all problem status',
    });

    panel.toolbar.insertItem(15, 'viewProblemStatus', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}


export default plugin;
