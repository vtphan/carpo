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

import { Cell, ICodeCellModel } from '@jupyterlab/cells';

// import { PanelLayout } from '@lumino/widgets';

// import { CellCheckButton, FloatingFeedbackWidget } from './widget';

import { FloatingFeedbackWidget } from './widget';


// import { CellInfo } from './model';

import { ISettingRegistry } from '@jupyterlab/settingregistry';

import { requestAPI } from './handler';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';

import {
  ToolbarButton,
  Dialog,
  showDialog,
  InputDialog,
  showErrorMessage,
  ICommandPalette
} from '@jupyterlab/apputils';

import { Widget } from '@lumino/widgets';




import { DocumentRegistry } from '@jupyterlab/docregistry';

import { SubmitCodeButton } from './share-code';
import { RaiseHandHelpButton } from './raise-hand-help';
// import { GetSolutionButton } from './get-solutions'
import { initializeNotifications, cleanupNotifications } from './sse-notifications';

import { LabIcon } from '@jupyterlab/ui-components';


const CommandIds = {
  /**
   * Command for carpo-student.
   */
  mainMenuRegister: 'jlab-carpo:main-register',
  mainMenuGetProblem: 'jlab-carpo:main-getProblem',
  mainMenuAbout: 'jlab-carpo:main-about',
  shareCodeCell: 'toolbar-button:share-code-cell'

};

export const fooIcon = new LabIcon({
  name: 'barpkg:foo',
  svgstr: `<svg fill="#000000" height="200px" width="200px" version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 488.9 488.9" xml:space="preserve"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <g> <path d="M411.448,100.9l-94.7-94.7c-4.2-4.2-9.4-6.2-14.6-6.2h-210.1c-11.4,0-20.8,9.4-20.8,20.8v330.8c0,11.4,9.4,20.8,20.8,20.8 h132.1v95.7c0,11.4,9.4,20.8,20.8,20.8s20.8-9.4,20.8-19.8v-96.6h132.1c11.4,0,19.8-9.4,19.8-19.8V115.5 C417.748,110.3,415.648,105.1,411.448,100.9z M324.048,70.4l39.3,38.9h-39.3V70.4z M378.148,331.9h-112.3v-82.8l17.7,16.3 c10,10,25,3.1,28.1-1c7.3-8.3,7.3-21.8-1-29.1l-52-47.9c-8.3-7.3-20.8-7.3-28.1,0l-52,47.9c-8.3,8.3-8.3,20.8-1,29.1 c8.3,8.3,20.8,8.3,29.1,1l17.7-16.3v82.8h-111.4V41.6h169.6v86.3c0,11.4,9.4,20.8,20.8,20.8h74.9v183.2H378.148z"></path> </g> </g></svg>`
});

/**
 * Initialization data for the carpo-student extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'carpo-student:plugin',
  autoStart: true,
  requires: [INotebookTracker, ICommandPalette],
  optional: [ISettingRegistry],
  activate: (
    app: JupyterFrontEnd,
    nbTrack: INotebookTracker,
    palette: ICommandPalette,
    settingRegistry: ISettingRegistry | null
  ) => {
    console.log('JupyterLab extension carpo-student is activated!');
    
    const { commands } = app;

    const cronTracker: Array<string> = [];
    const debounceTimers: Map<string, number> = new Map();
    const DEBOUNCE_DELAY = 15000; // 15 seconds delay after user stops typing

    // Attempt to register when extension is activated.
    requestAPI<any>('register', {
      method: 'POST',
      body: JSON.stringify({"serverUrl": "http://141.225.10.71:8081"})
    })
    .then(data => {
      console.log('Registration attempted:', data);
    })
    .catch(reason => {
      showErrorMessage('Registration Error', reason);
    });

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
          console.log('Snapshot sent.', data);
        }).catch(error => {
          console.error('Failed to send snapshot:', error);
        });
      }, DEBOUNCE_DELAY);

      debounceTimers.set(filename, newTimerId);

      initializeNotifications()
    };

    nbTrack.currentChanged.connect(() => {
      const currentNotebook = nbTrack.currentWidget;
      const notebook = currentNotebook.content;
      const filename = currentNotebook.context.path;
      const notebookTitle = notebook.title.label;

      // Disable if not inside Exercises directory
      if (!filename.includes('Exercises')) {
        return;
      }

      // Extract problem number from notebook filename for metadata
      const match = notebookTitle.match(/ex(\d+)\.ipynb/);
      const problemNumber = match ? parseInt(match[1]) : null;

      // Override cell creation methods for ex*.ipynb files - only once globally
      if (problemNumber !== null && !('_carpoOverrideApplied' in NotebookActions.insertBelow)) {
        
        // Store original functions
        const originalInsertBelow = NotebookActions.insertBelow;
        const originalInsertAbove = NotebookActions.insertAbove;

        // Override insertBelow method
        NotebookActions.insertBelow = (notebook: any) => {
          if (currentNotebook?.context.path.includes('Exercises') && 
          notebook.title.label.match(/ex\d+\.ipynb/)) {
            showErrorMessage('Carpo Error', `Cannot insert cell below on notebook ${notebook.title.label}.`);
            return;
          }
          return originalInsertBelow(notebook);
        };

        // Override insertAbove method
        NotebookActions.insertAbove = (notebook: any) => {
          if (currentNotebook?.context.path.includes('Exercises') && 
          notebook.title.label.match(/ex\d+\.ipynb/)) {
            showErrorMessage('Carpo Error', `Cannot insert cell above on notebook ${notebook.title.label}.`);
            return;
          }
          return originalInsertAbove(notebook);
        };

        // Mark that we've applied the override
        (NotebookActions.insertBelow as any)._carpoOverrideApplied = true;
        (NotebookActions.insertAbove as any)._carpoOverrideApplied = true;
      }

      // setup cron to capture code changes
      nbTrack.activeCellChanged.connect(() => {
        const cell: Cell = notebook.activeCell;
        const problem_id = cell.model.sharedModel.getMetadata("problem") || undefined;
        
        // Setup debounced snapshot sending when cell content changes
        if (Number.isInteger(problem_id) && cronTracker.indexOf(filename) === -1) {
          // Listen for changes to cell content
          cell.model.sharedModel.changed.connect(() => {
            sendDebouncedSnapshot(cell, filename, Number(problem_id));
          });
          cronTracker.push(filename);
        }
      })

      // kernel execution message
      NotebookActions.executed.connect(async (_, args) => {
          const { cell, notebook, success, error } = args;
          const content = cell.model.sharedModel.getSource()
          const problem_id = cell.model.sharedModel.getMetadata("problem") || undefined;

          if ( notebook.title.label.includes("ex") && !success) {
            const codeCellModel = cell.model as ICodeCellModel;
            const postBody = {
              message: JSON.stringify({
                "execution_count": `${codeCellModel.executionCount}`,
                "error_name": `${error.errorName}`,
                "error_value": `${error.errorValue}`
               }),
              code: content,
              problem_id: problem_id,
              snapshot: 4 // 1 is snapshot, 2 is submission, 3 is ask for help, 4 is code execution
            };
            requestAPI<any>('submissions', {
              method: 'POST',
              body: JSON.stringify(postBody)
            })
              .then(data => {
                console.log(data)
              })
              .catch(reason => {
                showErrorMessage('Code Run Error', reason);
                console.error(`Failed to run code.\n${reason}`);
              });
            };
      });

    });

    // Adds a command enabled only on ex__ notebooks
    commands.addCommand(CommandIds.shareCodeCell, {
      icon: fooIcon,
      caption: 'Share the content of this cell',
      execute: () => {
        // commands.execute('notebook:run-cell');
        const cell: Cell = nbTrack.currentWidget.content.activeCell;
        const content = cell.model.sharedModel.getSource()
        const problem_id = cell.model.sharedModel.getMetadata("problem") || undefined;

        if (problem_id === undefined ){
          showErrorMessage('Code Share Error', "Can not share non-exercise code cell.");
          return
        }
        
        const postBody = {
          message: "",
          code: content,
          problem_id: problem_id,
          snapshot: 2
        };
        requestAPI<any>('submissions', {
          method: 'POST',
          body: JSON.stringify(postBody)
        })
          .then(data => {
            if (data.msg === 'Submission saved successfully.') {
              data.msg = 'Code is sent to the instructor.';
            }
            showDialog({
              title: '',
              body: data.msg,
              buttons: [Dialog.okButton({ label: 'Ok' })]
            });
          })
          .catch(reason => {
            showErrorMessage('Code Share Error', reason);
            console.error(`Failed to share code to server.\n${reason}`);
          });
        
        initializeNotifications()
      },
      isVisible: () => nbTrack.currentWidget?.context.path.includes('ex')
    });

    const category = 'Extension Examples';
    const RegisterMenu = CommandIds.mainMenuRegister
    commands.addCommand(RegisterMenu, {
      label: 'Register',
      caption: 'Register to carpo',
      execute: (args: any) => {
        console.log("Args: ", args)
        InputDialog.getText({
          title: 'Enter server URL',
          label: 'Server URL:'
        }).then(value => {
          if (value.button.accept) {
            // console.log(`User entered: ${value.value}`);
            const reqData = {"serverUrl": `${value.value}`}
            requestAPI<any>('register', {
              method: 'POST',
              body: JSON.stringify(reqData)
            })
            .then(data => {
              console.log('Registration successful:', data);
              showDialog({
                title: 'Registration Successful',
                body: 'Student ' + data.name + ' has been registered.',
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });
            })
            .catch(reason => {
              showErrorMessage('Registration Error', reason);
              console.error(`Failed to register user.\n${reason}`);
            });
          }
        });
      }
    });

    // Add the command to the command palette
    palette.addItem({
      command: RegisterMenu,
      category: category,
      args: { origin: 'from the palette' }
    });

    const AboutMenu = CommandIds.mainMenuAbout
    commands.addCommand(AboutMenu, {
      label: 'About Carpo',
      caption: 'Carpo Information',
      execute: (args: any) => {
        const content = new Widget();
        content.node.innerHTML = `
          <h3>Use the following commands:</h3>
          <ol>
            <li><strong>Get Problems</strong>: Download active problems from the server.</li>
            <li><strong>AskForHelp</strong>: Request help with your code.</li>
            <li><strong>ViewFeedbacks</strong>: View feedbacks available to you.</li>
            <li><strong>GetSolution</strong>: Download the solution for the problem.</li>
          </ol>
          <p>Use the <em>Share</em> icon (1st button) in your cell to share your code.</p>
        `;

        showDialog({
          title: 'About Carpo',
          body: content,
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });
      }
    });

    // Add the command to the command palette
    palette.addItem({
      command: AboutMenu,
      category: category,
      args: { origin: 'from the palette' }
    });


    const GetProblemMenu = CommandIds.mainMenuGetProblem
    commands.addCommand(GetProblemMenu, {
      label: 'GetProblem',
      caption: 'Download Active Problem',
      execute: (args: any) => {
        requestAPI<any>('question', {
          method: 'GET'
        })
          .then(data => {
            // console.log(data);
            showDialog({
              title: 'Exercise Downloaded',
              body: data.msg,
              buttons: [Dialog.okButton({ label: 'Ok' })]
            });
          })
          .catch(reason => {
            showErrorMessage('Get Problem Error', reason);
            console.error(`Failed to get active questions.\n${reason}`);
          });
      }
    })

    //  tell the document registry about your widget extension:
    // app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    app.docRegistry.addWidgetExtension('Notebook', new SubmitCodeButton());
    app.docRegistry.addWidgetExtension('Notebook', new RaiseHandHelpButton());
    app.docRegistry.addWidgetExtension('Notebook', new ViewFeedbacksButton());
    app.docRegistry.addWidgetExtension('Notebook', new DownloadSolutionButton());
    
    // Add cleanup for notifications when the extension is deactivated
    // Note: JupyterFrontEnd doesn't have a disposed signal, so we'll handle cleanup
    // when the window is unloaded
    window.addEventListener('beforeunload', () => {
      cleanupNotifications();
    });
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

      requestAPI<any>('question', {
        method: 'GET'
      })
        .then(data => {
          // console.log(data);
          showDialog({
            title: 'Exercise Download',
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

    panel.toolbar.insertItem(10, 'getQuestion', button);
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
          // console.log(data);
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
      label: '💬 ViewFeedback',
      onClick: viewFeedbacks,
      tooltip: 'View feedback widget'
    });

    panel.toolbar.insertItem(12, 'viewFeedbacks', button);
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
      label: '💡 GetSolution',
      onClick: downloadSolution,
      tooltip: 'Download solution for this exercise'
    });

    panel.toolbar.insertItem(13, 'downloadSolution', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
