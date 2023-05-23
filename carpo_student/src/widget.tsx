import { ReactWidget } from '@jupyterlab/apputils';

import { Cell, CodeCell } from '@jupyterlab/cells';

import { CellInfo } from './model'

import { fileUploadIcon, LabIcon } from '@jupyterlab/ui-components';

import React from 'react';
import { requestAPI } from './handler';
import { Dialog, showDialog, showErrorMessage } from '@jupyterlab/apputils';


/**
 * 
 *
 * Note: A react component rendering a simple button with a jupyterlab icon
 *
 * @param icon - The subclass of LabIcon to show.
 * @param onClick - Method to call when the button is clicked.
 */
 interface IButtonComponent {
    icon: LabIcon;
    onClick: () => void;
  }
  
const ShareButton = ({
    icon,
    onClick
  }: IButtonComponent) => (
    <button
        type="button"
        onClick={() => onClick()}
        className="cellButton">
      <LabIcon.resolveReact
          icon={icon}
          className="cellButton-icon"
          tag="span"
          width="15px"
          height="15px"
      />
    </button>
  );

interface ICodeCellButtonComponent {
    cell: CodeCell;
    info: CellInfo;
}

const CodeCellButtonComponent = ({
    cell,
    info,
  }: ICodeCellButtonComponent): JSX.Element => {

  
    const shareCode = async () => {

        if (isNaN(info.problem_id)) {
            showErrorMessage('Code Share Error', "Invalid code block. Use specific problem notebook.");
            return
        }
        // clear message skeleton
        info.message = info.message.replace("## Message to instructor: ", "") 

        let postBody = {
            "message": info.message,
            "code": cell.model.value.text,
            "problem_id":info.problem_id,
            "snapshot": 2
        }
        console.log("From widget: ", postBody)
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
  
  
    return (
        <div>
            <ShareButton
                icon={fileUploadIcon}
                onClick={() => (shareCode)()}
            />

        </div>
      
    );
  };


export class CellCheckButton extends ReactWidget {
        cell: Cell = null;
        info: CellInfo = null;
      constructor(cell: Cell, info: CellInfo) {
          super();
          this.cell = cell;
          this.info = info;
          this.addClass('jp-CellButton');
      }
      render (): JSX.Element {
        return <CodeCellButtonComponent
            cell={this.cell as CodeCell}
            info = {this.info as CellInfo}
        />

    }

}
  