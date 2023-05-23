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
  CellCheckButton, FeedbackButton
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

import { GetSolutionButton } from './upload-solution'

/**
 * Initialization data for the carpo-teacher extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'carpo-teacher:plugin',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [IFileBrowserFactory],
  activate: (
      app: JupyterFrontEnd,
      nbTrack: INotebookTracker,
      browserFactory: IFileBrowserFactory | null,
      docManager: IDocumentManager,
      ) => {
    console.log('JupyterLab extension carpo-teacher is activated!');

    nbTrack.currentChanged.connect(() => {

      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;

      // If current Notebook is not inside Exercises/problem_ directory, disable all functionality.
      if (!nbTrack.currentWidget.context.path.includes("problem_")) {
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
          var sCell: Cell;
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
            code: cell.model.value.text
          };
          var header:string;

          // Get the status cell:
          notebook.widgets.map((c, index ) => {
            if (index == activeIndex+1) {
              sCell = c;
            }
          })

          header = cell.model.value.text.split("\n")[0]
          if(header.match(/^#[0-9]+ [0-9]+ [0-9]+$/)) {
            console.log("Submission Grading block.........")
            const newCheckButton: CellCheckButton = new CellCheckButton(cell,sCell,info);
  
            (cell.layout as PanelLayout).addWidget(newCheckButton);
            currentCell = cell;
            currentCellCheckButton = newCheckButton;

          } else {
            
            const newFeedbackButton: FeedbackButton = new FeedbackButton(cell,info);
            (cell.layout as PanelLayout).addWidget(newFeedbackButton);
            currentCell = cell;
            currentCellCheckButton = newFeedbackButton;

          }

        });

      });
    });
    
    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new RegisterButton());
    app.docRegistry.addWidgetExtension('Notebook', new GoToApp());
    app.docRegistry.addWidgetExtension('Notebook', new PublishProblemButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new ArchiveProblemButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new GetSolutionButton());
    
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
            title:'',
            body:  "Teacher "+ data.name + " has been registered.",
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
      label: 'Register',
      onClick: register,
      tooltip: 'Register as a Teacher',
    });

    panel.toolbar.insertItem(10, 'register', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class GoToApp implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const viewWebApp = () => {

    requestAPI<any>('view_app',{
      method: 'GET'
    })
      .then(data => {
        // console.log(data);
        window.open(
          data.url, "_blank");
      })
      .catch(reason => {
        showErrorMessage('View App Status Error', reason);
        console.error(
          `Failed to view app status.\n${reason}`
        );
      });

    };

    const button = new ToolbarButton({
      className: 'get-app-button',
      label: 'App',
      onClick: viewWebApp,
      tooltip: 'Go to the web app',
    });

    panel.toolbar.insertItem(11, 'viewWebApp', button);

    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}


export class NewSubmissionButtonExtension
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

          if (data.Remaining != 0 ){
            var msg = "Notebook " + data.sub_file + " is placed in folder Problem_"+ data.question +". There are " + data.remaining + " submissions in the queue."
          } else {
            var msg = "You have got 0 submissions. Please check again later.\n"
          }
          
          showDialog({
            title:'Submission Status',
            body: msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
          

          console.log(data)
    
        })
        .catch(reason => {
          showErrorMessage('Get Student Code Error', reason);
          console.error(
            `Failed to get student's code from the server. Please check your connection.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'sync-code-button',
      label: 'GetSubs',
      onClick: getSubmissions,
      tooltip: 'Download new submissions from students.',
    });

    panel.toolbar.insertItem(11, 'getStudentsCode', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class AllSubmissionButtonExtension
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
    const getGradedSubmissions = () => {
      NotebookActions.clearAllOutputs(panel.content);

      requestAPI<any>('graded_submissions',{
        method: 'GET'
      })
        .then(data => {
          
          showDialog({
            title:'',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
          

          console.log(data)
    
        })
        .catch(reason => {
          showErrorMessage('Get Graded Submissions Error', reason);
          console.error(
            `Failed to get student's code from the server. Please check your connection.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'sync-code-button',
      label: 'Graded',
      onClick: getGradedSubmissions,
      tooltip: 'Get all graded submissions.',
    });

    panel.toolbar.insertItem(12, 'getAllGradedSubmissions', button);
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
      var format:string
      var header:string
      var time_limit:string

      notebook.widgets.map((c:Cell, index:number) => {
        if (index === activeIndex ) {
          problem = c.model.value.text
          format = c.model.type
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


      header = problem.split('\n')[0]
      if(header.match(/[0-9]+[a-zA-Z]/)) {
        time_limit = header.match(/[0-9]+[a-zA-Z]/)[0]
      }


      let postBody = {
        "question": problem,
        "format": format,
        "time_limit": time_limit
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
      label: 'Publish',
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
      label: 'Unpublish',
      onClick: archiveProblem,
      tooltip: 'Unpublish the problem.',
    });

    panel.toolbar.insertItem(13, 'archivesProblem', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
