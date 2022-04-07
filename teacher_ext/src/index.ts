import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import { requestAPI } from './handler';

import { CellInfo } from './model'

import {
  NotebookActions,
  NotebookPanel,
  INotebookModel,
  INotebookTracker

} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { PanelLayout } from '@lumino/widgets';

import {
  CellCheckButton
} from './widget'

import { IFileBrowserFactory } from '@jupyterlab/filebrowser';

import {
  IDocumentManager
} from '@jupyterlab/docmanager';



// import { Cell } from '@jupyterlab/cells';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';
import { ToolbarButton, Dialog, showDialog,showErrorMessage } from '@jupyterlab/apputils';

// , InputDialog
import { DocumentRegistry } from '@jupyterlab/docregistry';


/**
 * Initialization data for the teacher-ext extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'teacher-ext:plugin',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [IFileBrowserFactory],
  activate: (
      app: JupyterFrontEnd,
      nbTrack: INotebookTracker,
      browserFactory: IFileBrowserFactory | null,
      docManager: IDocumentManager,
      ) => {
    console.log('JupyterLab extension teacher-ext is activated!');

    nbTrack.currentChanged.connect(() => {

      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;

      // If current Notebook is not inside Submissions directory, disable all functionality.
      if (!nbTrack.currentWidget.context.path.includes("Submissions/")) {
        return
      }


      notebookPanel.context.ready.then(async () => {

        let currentCell: Cell = null;
        let currentCellCheckButton: CellCheckButton = null;

        nbTrack.activeCellChanged.connect(() => {

          if (currentCell) {
            notebook.widgets.map((c: Cell) => {
              if (c.model.type == 'code' || c.model.type == 'markdown') {
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

          // const heading = cell.model.value.text.split("\n")[0].split(" ")
          const submission_id = function(text: string) {
            return Number(text.split("\n")[0].split(" ")[2])
          }

          const problem_id = function(text: string) {
            return Number(text.split("\n")[0].split(" ")[1])
          }

          const student_id = function(text: string) {
            return Number((text.split("\n")[0].split(" ")[0]).replace("#", ""))
          }

          var info : CellInfo = {
            id:  submission_id(cell.model.value.text),
            problem_id: problem_id(cell.model.value.text),
            student_id: student_id(cell.model.value.text),
            code: cell.model.value.text.split("\n")[1],
          };

          // For feedback case: cell is markdown so loop over the notebook widgets to get code cell before the active cell index
          if (cell.model.type == 'markdown' ){
            notebook.widgets.map((c,index) =>{
              if(index == activeIndex-1) {
                const code = c.model.value.text
                info.code = code  
                info.id = submission_id(code)
                info.student_id = student_id(code)
                info.problem_id = problem_id(code)
              }
            })
          }

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
    app.docRegistry.addWidgetExtension('Notebook', new ButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new PublishProblemButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new ArchiveProblemButtonExtension());
 
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

      requestAPI<any>('register',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title:'Register',
            body:  "User "+ data.name + " created as Teacher.",
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Registration Error', reason);
          console.error(
            `Failed to register user as Teacher.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'register-button',
      label: 'Setup Carpo',
      onClick: register,
      tooltip: 'Register as a Teacher',
    });

    panel.toolbar.insertItem(10, 'register', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}


export class ButtonExtension
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
    const getSubmissions = () => {
      NotebookActions.clearAllOutputs(panel.content);

      requestAPI<any>('submissions',{
        method: 'GET'
      })
        .then(data => {
          
          var msg = "You have got " + data.data.length + " submissions.\n Go to Notebooks inside Submissions directory."
        
          showDialog({
            title:'Submission Status',
            body: msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
          

          console.log(data)
    
        })
        .catch(reason => {
          showErrorMessage('Get recent submissions Error', reason);
          console.error(
            `Failed to get student's code from the server. Please check your connection.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'sync-code-button',
      label: 'Get Student Code',
      onClick: getSubmissions,
      tooltip: 'Get submissions from students.',
    });

    panel.toolbar.insertItem(11, 'getStudentsCode', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class PublishProblemButtonExtension
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
    const publishProblem = () => {
      NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;
      const activeIndex = notebook.activeCellIndex
      var problem:string

      notebook.widgets.map((c:Cell,index:number) => {
        if (index === activeIndex ) {
          problem = c.model.value.text
        }
      });

      if (problem.includes("#PID:")) {
        showErrorMessage('Publish Question Error', "Problem already published.")
        return
      }

      if (!problem) {
        showErrorMessage('Publish Question Error', "Problem is empty.")
        return
      }

      let postBody = {
        "question": problem
      }

      requestAPI<any>('problem',{
        method: 'POST',
        body: JSON.stringify(postBody)

      })
        .then(data => {
          console.log(data)
          notebook.widgets.map((c:Cell,index:number) => {
            if (index === activeIndex ) {
             c.model.value.text = "#PID:" + data.id + "\n" + problem
            }
          });


          showDialog({
          title:'New Questions Published',
          body: 'Problem ' + data.id + " is published.",
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });

        })
        .catch(reason => {
          showErrorMessage('Publish Question Error', reason);
          console.error(
            `Failed to publish question to the server.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'publish-problem-button',
      label: 'Publish Problem',
      onClick: publishProblem,
      tooltip: 'Publish New Problem.',
    });

    panel.toolbar.insertItem(12, 'publishNewProblem', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class ArchiveProblemButtonExtension
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
    const archiveProblem = () => {
      NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;
      const activeIndex = notebook.activeCellIndex
      var problem:string

      notebook.widgets.map((c:Cell,index:number) => {
        if (index === activeIndex ) {
          problem = c.model.value.text
        }
      });

      if (!problem.includes("#PID:")) {
        showErrorMessage('Unpublish Question Error', "Active problem not found.")
        return
      }

      var problem_id: number = parseInt((problem.split("\n")[0]).split("#PID:")[1]);

      let body = {
        "problem_id": problem_id
      }

      requestAPI<any>('problem',{
        method: 'DELETE',
        body: JSON.stringify(body)

      })
        .then(data => {
          console.log(data)
         
          showDialog({
          title:'Question Unpublished',
          body: 'Problem id ' + problem_id +' is  unpublished.',
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });

        })
        .catch(reason => {
          showErrorMessage('Unpublish Question Error', reason);
          console.error(
            `Failed to unpublish question.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'archive-problem-button',
      label: 'Unpublish Problem',
      onClick: archiveProblem,
      tooltip: 'Archive Problem.',
    });

    panel.toolbar.insertItem(13, 'archivesProblem', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
