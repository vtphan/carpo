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
import { CellInfo } from './model'

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
        const filename = panel.context.path
        const activeIndex = notebook.activeCellIndex

        var codeBlock:string

        var info : CellInfo = {
            problem_id: parseInt((filename.split("/").pop()).replace("ex","").replace(".ipynb",""))
          };

        notebook.widgets.map((c,index) =>{
            if (c.model.value.text.startsWith("## Message to instructor:")){
                info.message = c.model.value.text
            }
            if (index == activeIndex) {
                codeBlock = c.model.value.text
            }
        })

        if (!codeBlock.startsWith("## PID ")) {
            showErrorMessage('Code Share Error', "Invalid cell selected. Use a specific problem cell block.");
            return
          }


        let postBody = {
            "message": info.message,
            "code": codeBlock,
            "problem_id": info.problem_id,
            "snapshot": 1
        }

        console.log("Req body: ", postBody)
        requestAPI<any>('ask_for_help',{
            method: 'POST',
            body: JSON.stringify(postBody)
        })
        .then(data => {
            if (data.msg === "Submission saved successfully." ){
                if (info.message.length > 27) {
                    data.msg = 'Code & message is sent to the instructor.'
                } else {
                    data.msg = 'Code is sent to the instructor.'
                }
            }
            showDialog({
                title:'',
                body: data.msg,
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });
        })
        .catch(reason => {
            showErrorMessage('Code Share Error', reason);
            console.error(
            `Failed to share code to server.\n${reason}`
            );
        });

        // Put on watch list

    };

    const button = new ToolbarButton({
      className: 'raise-hand-button',
      label: 'AskForHelp',
      onClick: raiseHand,
      tooltip: 'Ask the instructor to help you.',
    });

    panel.toolbar.insertItem(12, 'AskForHelp', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}