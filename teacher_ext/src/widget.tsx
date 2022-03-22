import { ReactWidget } from '@jupyterlab/apputils';

import { Cell, CodeCell } from '@jupyterlab/cells';

import { CellInfo } from './model'

import { checkIcon, closeIcon,LabIcon,saveIcon } from '@jupyterlab/ui-components';

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
  
const GradeButton = ({
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

const SendButton = ({
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
  
    const submitGrade = async (val: Boolean) => {

        console.log("From widget: ", info)
        let postBody = {
            "student_id": info.student_id,
            "submission_id": info.id,
            "problem_id": info.problem_id,
            "score": val ? 1 : 2
        }
        var status : string = val ? "Correct.": "Incorrect." 

        console.log("Grade: ", postBody)
     
        requestAPI<any>('submissions/grade',{
            method: 'POST',
            body: JSON.stringify(postBody)
        }).then(data => {
            var msg = "This submission is now graded as " + status
            showDialog({
                title:'Grading Status',
                body: msg,
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });
            })
            .catch(reason => {
            showErrorMessage('Submission Grade Error', reason);
            console.error(
                `Failed to grade the submission. \n${reason}`
            );
        });


    };
  
  
    return (
        <div>
            <GradeButton
                icon={checkIcon}
                onClick={() => (submitGrade)(true)}
            />
            <GradeButton
                icon={closeIcon}
                onClick={() => (submitGrade)(false)}
            />


        </div>
      
    );
  };

const MarkdownCellButtonComponent = ({
    cell,
    info,
}: ICodeCellButtonComponent): JSX.Element => {

    const sendFeedback =async () => {
        let postBody = {
            "student_id": info.student_id,
            "submission_id": info.id,
            "problem_id": info.problem_id,
            "code": info.code,
            "message": info.message,
            "comment": cell.model.value.text
        }

        console.log("Feedback: ", postBody)
    
        requestAPI<any>('submissions/feedbacks',{
            method: 'POST',
            body: JSON.stringify(postBody)
        }).then(data => {

            showDialog({
                title:'Feedback Status',
                body: "Feedback is now provided.",
                buttons: [Dialog.okButton({ label: 'Ok' })]
              });

            })
            .catch(reason => {
            showErrorMessage('Submission Feedback Error', reason);
            console.error(
                `Failed to save feedback. \n${reason}`
            );
        });
        
    }

    return (
        <SendButton 
            icon={saveIcon}
            onClick={() => (sendFeedback)()}
        />
    )
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
            
            case 'markdown':
            return <MarkdownCellButtonComponent
                cell={this.cell as CodeCell}
                info = {this.info as CellInfo}
            />

            default:
                break;
        }

    }

}
  