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
 * Initialization data for the jupyterlab-student-ext extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-student-ext:plugin',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [ISettingRegistry],
  activate: (
      app: JupyterFrontEnd, 
      nbTrack: INotebookTracker,
      settingRegistry: ISettingRegistry | null
    ) => {
    console.log('JupyterLab extension jupyterlab-student-ext is activated!');

    nbTrack.currentChanged.connect(() => {

      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;
      const filename = notebookPanel.context.path

      // Disable Code Share functionality if not the "carpo-problem-"" Notebook.
      if (!filename.includes("carpo-problem-")) {
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
            code: cell.model.value.text,
            problem_id: parseInt((filename.split("-").pop()).replace(".ipynb",""))
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
    app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    // app.docRegistry.addWidgetExtension('Notebook', new CodeSubmissionButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetFeedbackButton());

  }
};

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
            title:'New Question Downloaded',
            body: 'A new problem has been downloaded to ' + data.msg +" file.",
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Get Question Error', reason);
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-question-button',
      label: 'Get Question',
      onClick: getQuestion,
      tooltip: 'Get Latest Question from Server',
    });

    panel.toolbar.insertItem(10, 'getQuestion', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class CodeSubmissionButton
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
    const sendCode = () => {
      // NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;

      const activeIndex = notebook.activeCellIndex
      var message:string, code :string

      notebook.widgets.map((c:Cell,index:number) => {
        // This is Markdown cell
        if (index === activeIndex-1 ) {
          message = c.model.value.text

        }
        // This is Code cell & Active cell
        if (index === activeIndex) {
          code = c.model.value.text
        }
      });

      const filename = panel.context.path


      let postBody = {
        "message": message,
        "code": code,
        "problem_id":(filename.split("-").pop()).replace(".ipynb","")
      }

      console.log("body: ",postBody)

      requestAPI<any>('submissions',{
        method: 'POST',
        body: JSON.stringify(postBody)
      })
        .then(data => {
          console.log(data);
          window.alert(
            `This code has been shared to the server.`
          );
        })
        .catch(reason => {
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'send-code-button',
      label: 'Send All Code',
      onClick: sendCode,
      tooltip: 'Send code to Go Server',
    });

    panel.toolbar.insertItem(11, 'sendCodes', button);
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
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-feedback-button',
      label: 'Get Feedback',
      onClick: getFeedback,
      tooltip: 'Get Feedback to your Submission',
    });

    panel.toolbar.insertItem(11, 'getFeedback', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
