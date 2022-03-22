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

        let postBody = {
            "message": info.message,
            "code": cell.model.value.text,
            "problem_id":info.problem_id
        }
        
        console.log("From widget: ", postBody)

        requestAPI<any>('submissions',{
            method: 'POST',
            body: JSON.stringify(postBody)
        })
        .then(data => {
            console.log(data);
            showDialog({
                title:'Code Share',
                body: 'This code has been shared to the server.',
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });

        })
        .catch(reason => {
            showErrorMessage('Code Share Error', reason);
            console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
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

        switch(this.cell.model.type) {
            case 'code':
                return <CodeCellButtonComponent
                    cell={this.cell as CodeCell}
                    info = {this.info as CellInfo}
                />

            default:
                break;
        }

    }

}
  