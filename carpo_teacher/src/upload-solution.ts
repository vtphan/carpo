import {
    // INotebookTracker,
    // NotebookActions,
    NotebookPanel,
    INotebookModel,
  
  } from '@jupyterlab/notebook';
import { Cell } from '@jupyterlab/cells';
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

    const uploadSolution = () => {
        const notebook = panel.content;
        const activeIndex = notebook.activeCellIndex
        var code_block:string
        var solution: string

        notebook.widgets.map((c:Cell,index:number) => {
            if (index === activeIndex ) {
                code_block = c.model.sharedModel.getSource()
            }
        });

        if (!code_block.includes("#PID:")) {
            showErrorMessage('Upload Solution Error', "Active problem not found.")
            return
        }

        var problem_id: number = parseInt((code_block.split("\n")[0]).split("#PID:")[1]);

        solution = code_block.split('\n').slice(1).join('\n').trim();

        let body = {
            "problem_id": problem_id,
            "code": solution
        }
        

        requestAPI<any>('solution',{
            method: 'POST',
            body: JSON.stringify(body)
          })
            .then(data => {
              showDialog({
                title:'Solution Uploaded',
                body: 'Solution uploaded for ProblemID ' + problem_id +'.',
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });
      
            })
            .catch(reason => {
            showErrorMessage('Upload Solution Error', reason);
            console.error(
                    `Failed to upload solution to problem.\n${reason}`
                );
            });

    };

    const button = new ToolbarButton({
      className: 'upload-solution-button',
      label: 'UploadSolution',
      onClick: uploadSolution,
      tooltip: 'Upload solutions to the problem.',
    });

    panel.toolbar.insertItem(12, 'getSolutions', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}