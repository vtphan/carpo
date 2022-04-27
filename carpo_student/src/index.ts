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

    nbTrack.currentChanged.connect(() => {

      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;
      const filename = notebookPanel.context.path

      // Disable Code Share functionality if inside Feedback directory
      if (filename.includes("Feedback")){
        return 
      }

      // Disable if not inside Carpo directory
      if (!filename.includes("Carpo")) {
        return
      }

      notebookPanel.context.ready.then(async () => {

        let currentCell: Cell = null;
        let currentCellCheckButton: CellCheckButton = null;

        nbTrack.activeCellChanged.connect(() => {

          if (currentCell) {
            notebook.widgets.map((c: Cell) => {
              if (c.model.type == 'code') {
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
            problem_id: parseInt((filename.split("/").pop()).replace("p","").replace(".ipynb",""))
          };

          // Get the message block referencing the active cell.
          notebook.widgets.map((c,index) =>{
            if(index == activeIndex-1) {
              info.message = c.model.value.text
            }
          })
       

          const newCheckButton: CellCheckButton = new CellCheckButton(
            cell,info);

          (cell.layout as PanelLayout).addWidget(newCheckButton);

          // Set the current cell and button for future
          // reference
          currentCell = cell;
          currentCellCheckButton = newCheckButton;

        });

      });
    });


    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new RegisterButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetFeedbackButton());
    app.docRegistry.addWidgetExtension('Notebook', new ViewSubmissionStatusButton());


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
            title:'Register',
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
      label: 'Setup Carpo',
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
            title:'Questions',
            body:  data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Get Question Error', reason);
          console.error(
            `Failed to get active questions.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-question-button',
      label: 'Get Question',
      onClick: getQuestion,
      tooltip: 'Get Latest Question from Server',
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
            title:'Teacher Feedback',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
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
      label: 'Get Feedback',
      onClick: getFeedback,
      tooltip: 'Get Feedback to your Submission',
    });

    panel.toolbar.insertItem(12, 'getFeedback', button);
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
      label: 'Submission Status',
      onClick: viewStatus,
      tooltip: 'View your submissions status',
    });

    panel.toolbar.insertItem(13, 'viewStatus', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}


export default plugin;
