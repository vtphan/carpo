import {
  // INotebookTracker,
  // NotebookActions,
  NotebookPanel,
  INotebookModel
} from '@jupyterlab/notebook';
import { DocumentRegistry } from '@jupyterlab/docregistry';
import { IDisposable, DisposableDelegate } from '@lumino/disposable';
import {
  ToolbarButton,
  Dialog,
  showDialog,
  showErrorMessage
} from '@jupyterlab/apputils';

import { requestAPI } from './handler';
import { Cell } from '@jupyterlab/cells';


import { initializeNotifications } from './sse-notifications';


export class SubmitCodeButton
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
    const shareCode = () => {
      const notebook = panel.content;

      const cell: Cell = notebook.activeCell;
      const content = cell.model.sharedModel.getSource()
      const problem_id = cell.model.sharedModel.getMetadata("problem") || undefined;

      let codeBlock: string | undefined;
      let pID: number | undefined

      // Search for problem cell in the notebook.
      // Allows student to submit code independent of activeCell
      notebook.widgets.map((c) => {
        const cellMetadata = c.model.sharedModel.getMetadata();
        if (cellMetadata['problem'] !== undefined) {
          codeBlock = c.model.sharedModel.getSource()
          pID = Number(cellMetadata['problem'])
        }
      })

      const postBody = {
        message: "",
        code: codeBlock ?? content,
        problem_id: pID ?? problem_id,
        snapshot: 2
      };

      // console.log('Req body: ', postBody);
      requestAPI<any>('submissions', {
        method: 'POST',
        body: JSON.stringify(postBody)
      })
        .then(data => {
          if (data.msg === 'Submission saved successfully.') {
            data.msg = 'Code is submitted.';
          }
          showDialog({
            title: '',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
          
        })
        .catch(reason => {
          showErrorMessage('Code Submit Error', reason);
          console.error(`Failed to submit code.\n${reason}`);
        });

        initializeNotifications()
    };

    const button = new ToolbarButton({
      className: 'submit-code-button',
      label: '➤ Submit',
      onClick: shareCode,
      tooltip: 'Submit your code'
    });

    panel.toolbar.insertItem(10, 'shareCode', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}
