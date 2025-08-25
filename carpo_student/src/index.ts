import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import {
  INotebookTracker,
  NotebookActions,
  NotebookPanel,
  INotebookModel
} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { PanelLayout } from '@lumino/widgets';

import { CellCheckButton, FloatingFeedbackWidget } from './widget';

import { CellInfo } from './model';

import { ISettingRegistry } from '@jupyterlab/settingregistry';

import { requestAPI } from './handler';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';

import {
  ToolbarButton,
  Dialog,
  showDialog,
  showErrorMessage
} from '@jupyterlab/apputils';

import { DocumentRegistry } from '@jupyterlab/docregistry';

import { ShareCodeButton } from './share-code';
import { RaiseHandHelpButton } from './raise-hand-help';
// import { GetSolutionButton } from './get-solutions'
import { initializeNotifications, cleanupNotifications } from './sse-notifications';

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
    
    // Initialize SSE notifications
    // initializeNotifications();
    
    const cronTracker: Array<string> = [];
    const debounceTimers: Map<string, number> = new Map();
    const DEBOUNCE_DELAY = 10000; // 10 seconds delay after user stops typing

    // Debounced function to send code snapshot
    const sendDebouncedSnapshot = (cell: Cell, filename: string, problemId: number) => {
      const timerId = debounceTimers.get(filename);
      if (timerId) {
        clearTimeout(timerId);
      }

      const newTimerId = window.setTimeout(() => {
        const postBody = {
          message: '',
          code: cell.model.sharedModel.getSource(),
          problem_id: problemId,
          snapshot: 1
        };
        requestAPI<any>('submissions', {
          method: 'POST',
          body: JSON.stringify(postBody)
        }).then(data => {
          console.log('Snapshot sent (debounced).', data);
        }).catch(error => {
          console.error('Failed to send snapshot:', error);
        });
      }, DEBOUNCE_DELAY);

      debounceTimers.set(filename, newTimerId);

      initializeNotifications()
    };

    nbTrack.currentChanged.connect(() => {
      // console.log("my tracker: ", tracker);
      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;
      const filename = notebookPanel.context.path;

      // Disable if not inside Exercises directory
      if (!filename.includes('Exercises')) {
        return;
      }

      notebookPanel.context.ready.then(async () => {
        let currentCell: Cell = null;
        let currentCellCheckButton: CellCheckButton = null;

        nbTrack.activeCellChanged.connect(() => {
          let question: string;

          if (currentCell) {
            notebook.widgets.map((c: Cell) => {
              if (c.model.type === 'code' || c.model.type === 'markdown') {
                const currentLayout = c.layout as PanelLayout;
                currentLayout.widgets.map(w => {
                  if (w === currentCellCheckButton) {
                    currentLayout.removeWidget(w);
                  }
                });
              }
            });
          }

          const cell: Cell = notebook.activeCell;
          const activeIndex = notebook.activeCellIndex;

          const info: CellInfo = {
            problem_id: parseInt(
              filename.split('/').pop().replace('ex', '').replace('.ipynb', '')
            )
          };

          // Get the message block referencing the active cell.
          notebook.widgets.map((c, index) => {
            // if (c.model.toJSON().source[0].startsWith('## Message to instructor:')) {
            //   info.message = c.model.value.text;
            // }
            if (index === activeIndex) {
              question = c.model.sharedModel.getSource()
              if (question.includes('## PID ')) {
                const newCheckButton: CellCheckButton = new CellCheckButton(
                  cell,
                  info
                );
                (cell.layout as PanelLayout).addWidget(newCheckButton);
                currentCellCheckButton = newCheckButton;

                // Setup debounced snapshot sending when cell content changes
                if (cronTracker.indexOf(filename) === -1) {
                  // Listen for changes to cell content
                  c.model.sharedModel.changed.connect(() => {
                    sendDebouncedSnapshot(c, filename, info.problem_id);
                  });
                  cronTracker.push(filename);
                }
              }
            }
          });

          currentCell = cell;
        });
      });
    });

    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new RegisterButton());
    app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    app.docRegistry.addWidgetExtension('Notebook', new RaiseHandHelpButton());
    // app.docRegistry.addWidgetExtension('Notebook', new ViewSubmissionStatusButton());
    app.docRegistry.addWidgetExtension('Notebook', new ViewFeedbacksButton());
    app.docRegistry.addWidgetExtension('Notebook', new DownloadSolutionButton());
    app.docRegistry.addWidgetExtension('Notebook', new ShareCodeButton());
    // app.docRegistry.addWidgetExtension('Notebook', new viewProblemStatusExtension());
    
    // Add cleanup for notifications when the extension is deactivated
    // Note: JupyterFrontEnd doesn't have a disposed signal, so we'll handle cleanup
    // when the window is unloaded
    window.addEventListener('beforeunload', () => {
      cleanupNotifications();
    });
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

      requestAPI<any>('register', {
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title: '',
            body: 'Student ' + data.name + ' has been registered.',
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
        })
        .catch(reason => {
          showErrorMessage('Registration Error', reason);
          console.error(`Failed to register user as Student.\n${reason}`);
        });
    };

    const button = new ToolbarButton({
      className: 'register-button',
      label: 'Register',
      onClick: register,
      tooltip: 'Register as a Student'
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

      requestAPI<any>('question', {
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title: '',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
        })
        .catch(reason => {
          showErrorMessage('Get Problem Error', reason);
          console.error(`Failed to get active questions.\n${reason}`);
        });
    };

    const button = new ToolbarButton({
      className: 'get-question-button',
      label: 'GetProblem',
      onClick: getQuestion,
      tooltip: 'Get Latest Problem From Server'
    });

    panel.toolbar.insertItem(11, 'getQuestion', button);
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
      requestAPI<any>('view_student_status', {
        method: 'GET'
      })
        .then(data => {
          console.log(data);
          window.open(data.url, '_blank');
        })
        .catch(reason => {
          showErrorMessage('View Status Error', reason);
          console.error(`Failed to view student submission status.\n${reason}`);
        });
    };

    const button = new ToolbarButton({
      className: 'get-status-button',
      label: 'Status',
      onClick: viewStatus,
      tooltip: 'View your submissions status'
    });

    panel.toolbar.insertItem(13, 'viewStatus', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

// Currently disabled
export class ViewFeedbacksButton
  implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  private static feedbackWidgets = new Map<string, FloatingFeedbackWidget>();

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
    // Get the notebook filename as unique identifier
    const filename = context.path;
    const notebookName = filename.split('/').pop() || '';
    
    // Only show the feedback button for notebooks starting with 'ex'
    if (!notebookName.startsWith('ex')) {
      // Return an empty disposable for non-exercise notebooks
      return new DisposableDelegate(() => {});
    }

    const viewFeedbacks = () => {
      // Check if feedback widget already exists for this filename
      let floatingFeedback = ViewFeedbacksButton.feedbackWidgets.get(filename);
      
      if (!floatingFeedback) {
        // Create new feedback widget if it doesn't exist
        floatingFeedback = new FloatingFeedbackWidget(filename);
        ViewFeedbacksButton.feedbackWidgets.set(filename, floatingFeedback);
        
        // Add cleanup when widget is closed
        const originalClose = floatingFeedback.close.bind(floatingFeedback);
        floatingFeedback.close = () => {
          originalClose();
          ViewFeedbacksButton.feedbackWidgets.delete(filename);
        };
      }
      
      floatingFeedback.show();
    };

    // Setup cleanup when notebook panel is disposed
    const cleanupFeedback = () => {
      const filename = context.path;
      const widget = ViewFeedbacksButton.feedbackWidgets.get(filename);
      if (widget) {
        widget.close();
        ViewFeedbacksButton.feedbackWidgets.delete(filename);
      }
    };

    // Listen for panel disposal
    panel.disposed.connect(cleanupFeedback);

    const button = new ToolbarButton({
      className: 'view-feedbacks-button',
      label: 'ViewFeedbacks',
      onClick: viewFeedbacks,
      tooltip: 'View feedback widget'
    });

    panel.toolbar.insertItem(13, 'viewFeedbacks', button);
    return new DisposableDelegate(() => {
      button.dispose();
      // Clean up feedback widget when button is disposed
      const filename = context.path;
      const widget = ViewFeedbacksButton.feedbackWidgets.get(filename);
      if (widget) {
        widget.close();
        ViewFeedbacksButton.feedbackWidgets.delete(filename);
      }
    });
  }
}

// export class viewProblemStatusExtension
//   implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
// {
//   /**
//    * Create a new extension for the notebook panel widget.
//    *
//    * @param panel Notebook panel
//    * @param context Notebook context
//    * @returns Disposable on the added button
//    */
//   createNew(
//     panel: NotebookPanel,
//     context: DocumentRegistry.IContext<INotebookModel>
//   ): IDisposable {
//     const viewProblemStatus = () => {
//       requestAPI<any>('view_problem_list', {
//         method: 'GET'
//       })
//         .then(data => {
//           console.log(data);
//           window.open(data.url, '_blank');
//         })
//         .catch(reason => {
//           showErrorMessage('View Problem Status Error', reason);
//           console.error(`Failed to view problem status.\n${reason}`);
//         });
//     };

//     const button = new ToolbarButton({
//       className: 'get-status-button',
//       label: 'Problems',
//       onClick: viewProblemStatus,
//       tooltip: 'View all problem status'
//     });

//     panel.toolbar.insertItem(15, 'viewProblemStatus', button);
//     return new DisposableDelegate(() => {
//       button.dispose();
//     });
//   }
// }

export class DownloadSolutionButton
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
    const downloadSolution = () => {
      const notebook = panel.content;
      const filename = context.path;
      
      // Only show for exercise notebooks
      if (!filename.includes('Exercises') || !filename.includes('ex')) {
        showErrorMessage('Invalid Notebook', 'Solution download is only available for exercise notebooks.');
        return;
      }

      // Extract problem_id from filename (e.g., ex001.ipynb -> 1)
      const match = filename.match(/ex(\d+)\.ipynb/);
      if (!match) {
        showErrorMessage('Invalid Filename', 'Cannot extract problem ID from notebook filename.');
        return;
      }

      const problemId = parseInt(match[1]);
      
      requestAPI<any>(`solutions/problem/${problemId}`, {
        method: 'GET'
      })
        .then(data => {
          
          if (data.data.code) {
            // Create a new code cell with the solution
            const solutionCode = `# Solution for Problem ${problemId}\n${data.data.code}`;
            
            // Move to the last cell first
            notebook.activeCellIndex = notebook.widgets.length - 1;
            
            // Insert a new code cell at the end of the notebook
            NotebookActions.insertBelow(notebook);
            
            // Get the newly created cell (should be the last cell now)
            const activeCell = notebook.activeCell;
            if (activeCell && activeCell.model.type === 'code') {
              // Set the source code
              activeCell.model.sharedModel.setSource(solutionCode);
            }
            
            // Scroll to the new cell
            notebook.scrollToItem(notebook.widgets.length - 1);
            
            showDialog({
              title: 'Solution Downloaded',
              body: `Solution for Problem ${problemId} has been added to your notebook.`,
              buttons: [Dialog.okButton({ label: 'Ok' })]
            });
          } else {
            showDialog({
              title: 'No Solution', 
              body: 'No solution available for this problem.',
              buttons: [Dialog.okButton({ label: 'Ok' })]
            });
          }
        })
        .catch(reason => {
          showErrorMessage('Download Solution Error', reason);
          console.error(`Failed to download solution.\n${reason}`);
        });
    };

    const button = new ToolbarButton({
      className: 'download-solution-button',
      label: 'GetSolution',
      onClick: downloadSolution,
      tooltip: 'Download solution for this exercise'
    });

    panel.toolbar.insertItem(14, 'downloadSolution', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
