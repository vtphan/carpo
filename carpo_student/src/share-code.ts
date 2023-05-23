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
            "snapshot": 2
        }

        console.log("Req body: ", postBody)
        requestAPI<any>('submissions',{
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

            // Keep checking for new feedback.
            // This setInterval will be cleared once the feedback is downloaded (after reload())
            // setInterval(function(){
            //     // console.log("Checking for feedback...")
            //     requestAPI<any>('feedback',{
            //     method: 'GET'
            //     })
            //     .then(data => {
            //         // console.log(data);
            //         if (data['hard-reload'] != -1) {
            //         showDialog({
            //             title:'',
            //             body: data.msg,
            //             buttons: [Dialog.okButton({ label: 'Ok' })]
            //         }).then( result => {
            //             if (result.button.accept ) {
            //                 window.location.reload();
            //             }
            //         })
        
            //         }
                    
            //     })
            //     .catch(reason => {
            //         showErrorMessage('Get Feedback Error', reason);
            //         console.error(
            //         `Failed to fetch recent feedbacks.\n${reason}`
            //         );
            //     });
        
            // }, 60000);

        })
        .catch(reason => {
            showErrorMessage('Code Share Error', reason);
            console.error(
            `Failed to share code to server.\n${reason}`
            );
        });

    };

    const button = new ToolbarButton({
      className: 'share-code-button',
      label: 'ShareCode',
      onClick: shareCode,
      tooltip: 'Share your code to the instructor.',
    });

    panel.toolbar.insertItem(12, 'shareCode', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}