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
import { CellInfo } from './model';

import { initializeNotifications } from './sse-notifications';


export class ShareCodeButton
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
      const filename = panel.context.path;
      const activeIndex = notebook.activeCellIndex;

      let codeBlock: string;

      const info: CellInfo = {
        problem_id: parseInt(
          filename.split('/').pop().replace('ex', '').replace('.ipynb', '')
        )
      };

      notebook.widgets.map((c, index) => {
        // if (c.model.toJSON().source[0].text.startsWith('## Message to instructor:')) {
        //   info.message = c.model.value.text;
        // }
        if (index === activeIndex) {
          // codeBlock = c.model.toJSON().source[0];
          codeBlock = c.model.sharedModel.getSource()
          console.log("content: ", codeBlock)
        }
      });

      if (!codeBlock.startsWith('## PID ')) {
        showErrorMessage(
          'Code Share Error',
          'Invalid cell selected. Use a specific problem cell block.'
        );
        return;
      }

      const postBody = {
        message: info.message,
        code: codeBlock,
        problem_id: info.problem_id,
        snapshot: 2
      };

      // console.log('Req body: ', postBody);
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
    };

    const button = new ToolbarButton({
      className: 'share-code-button',
      label: 'ShareCode',
      onClick: shareCode,
      tooltip: 'Share your code to the instructor.'
    });

    panel.toolbar.insertItem(15, 'shareCode', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}
