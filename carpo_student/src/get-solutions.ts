import {
    // INotebookTracker,
    // NotebookActions,
    NotebookPanel,
    INotebookModel,
  
  } from '@jupyterlab/notebook';
import { DocumentRegistry } from '@jupyterlab/docregistry';
import { IDisposable, DisposableDelegate } from '@lumino/disposable';
import { ToolbarButton,Dialog, showDialog,showErrorMessage } from '@jupyterlab/apputils';

import { requestAPI } from './handler';

export class GetSolutionButton
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

    const getSolutions = () => {

        requestAPI<any>('solution',{
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
              showErrorMessage('Get Solution Error', reason);
              console.error(
                `Failed to get problem solutions.\n${reason}`
              );
            });

    };

    const button = new ToolbarButton({
      className: 'get-solution-button',
      label: 'GetSolutions',
      onClick: getSolutions,
      tooltip: 'Download solutions to problems.',
    });

    panel.toolbar.insertItem(14, 'getSolutions', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}