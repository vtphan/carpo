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

// import { CellInfo } from './model';

export class RaiseHandHelpButton
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
    const raiseHand = () => {
      const notebook = panel.content;
      // const filename = panel.context.path;
      // const activeIndex = notebook.activeCellIndex;

      const cell: Cell = notebook.activeCell;
      const content = cell.model.sharedModel.getSource()
      const problem_id = cell.model.sharedModel.getMetadata("problem") || undefined;

      let codeBlock: string | undefined;
      let pID: number | undefined

      // Search for problem cell in the notebook.
      // Allows student to request help independent of activeCell
      notebook.widgets.map((c) => {
        const cellMetadata = c.model.sharedModel.getMetadata();
        if (cellMetadata['problem'] !== undefined) {
          codeBlock = c.model.sharedModel.getSource()
          pID = Number(cellMetadata['problem'])
        }
      })


      const postBody = {
        message: '',
        code: codeBlock ?? content,
        problem_id: pID ?? problem_id,
        snapshot: 3  // 1 is snapshot, 2 is submission, 3 is ask for help,
      };

      // console.log('Req body: ', postBody);
      requestAPI<any>('ask_for_help', {
        method: 'POST',
        body: JSON.stringify(postBody)
      })
        .then(data => {
          if (data.msg === 'Submission saved successfully.') {
              data.msg = 'Code is shared and you will get feedback soon.';
          }
          showDialog({
            title: 'Help Request Sent',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
        })
        .catch(reason => {
          showErrorMessage('Code Share Error', reason);
          console.error(`Failed to share code to server.\n${reason}`);
        });

    };

    const button = new ToolbarButton({
      className: 'raise-hand-button',
      label: '❓ AskForHelp',
      onClick: raiseHand,
      tooltip: 'Request help from instructor'
    });

    panel.toolbar.insertItem(11, 'AskForHelp', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}
