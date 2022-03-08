import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import {
  // INotebookTracker,
  // NotebookActions,
  NotebookPanel,
  INotebookModel,

} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { ISettingRegistry } from '@jupyterlab/settingregistry';

import { requestAPI } from './handler';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';

import { ToolbarButton,Dialog, showDialog } from '@jupyterlab/apputils';

import { DocumentRegistry } from '@jupyterlab/docregistry';


/**
 * Initialization data for the jupyterlab-student-ext extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-student-ext:plugin',
  autoStart: true,
  optional: [ISettingRegistry],
  activate: (app: JupyterFrontEnd, settingRegistry: ISettingRegistry | null) => {
    console.log('JupyterLab extension jupyterlab-student-ext is activated!');

    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new GetQuestionButton());
    app.docRegistry.addWidgetExtension('Notebook', new CodeSubmissionButton());
    app.docRegistry.addWidgetExtension('Notebook', new SubmissionFeedbackButton());

  }
};

export class GetQuestionButton
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
    const getQuestion = () => {
      // NotebookActions.clearAllOutputs(panel.content);

      // const notebook = panel.content;

      requestAPI<any>('question',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title:'New Question Downloaded',
            body: 'A new problem has been downloaded to ' + data.msg +" file.",
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-question-button',
      label: 'Get Question',
      onClick: getQuestion,
      tooltip: 'Get Latest Question from Server',
    });

    panel.toolbar.insertItem(10, 'getQuestion', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class CodeSubmissionButton
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
    const sendCode = () => {
      // NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;

      const activeIndex = notebook.activeCellIndex
      var message:string, code :string

      notebook.widgets.map((c:Cell,index:number) => {
        // This is Markdown cell
        if (index === activeIndex-1 ) {
          message = c.model.value.text

        }
        // This is Code cell & Active cell
        if (index === activeIndex) {
          code = c.model.value.text
        }
      });

      const filename = panel.context.path


      let postBody = {
        "message": message,
        "code": code,
        "problem_id":(filename.split("-").pop()).replace(".ipynb","")
      }

      console.log("body: ",postBody)

      requestAPI<any>('submissions',{
        method: 'POST',
        body: JSON.stringify(postBody)
      })
        .then(data => {
          console.log(data);
          window.alert(
            `This code has been shared to the server.`
          );
        })
        .catch(reason => {
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'send-code-button',
      label: 'Send All Code',
      onClick: sendCode,
      tooltip: 'Send code to Go Server',
    });

    panel.toolbar.insertItem(11, 'sendCodes', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class SubmissionFeedbackButton
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
    const getFeedback = () => {

      requestAPI<any>('feedback',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);
          showDialog({
            title:'Teacher Feedback',
            body: data.msg,
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
        })
        .catch(reason => {
          console.error(
            `The student_ext server extension appears to be missing.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'get-feedback-button',
      label: 'Get Feedback',
      onClick: getFeedback,
      tooltip: 'Get Feedback to your Submission',
    });

    panel.toolbar.insertItem(12, 'getFeedback', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
