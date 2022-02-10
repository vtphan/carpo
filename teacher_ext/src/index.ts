import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import { requestAPI } from './handler';

import { CellInfo } from './model'

import {
  NotebookActions,
  NotebookPanel,
  INotebookModel,
  INotebookTracker

} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { PanelLayout } from '@lumino/widgets';

import {
  CellCheckButton
} from './widget'

import { IFileBrowserFactory } from '@jupyterlab/filebrowser';

import {
  IDocumentManager
} from '@jupyterlab/docmanager';



// import { Cell } from '@jupyterlab/cells';

import { IDisposable, DisposableDelegate } from '@lumino/disposable';
import { ToolbarButton } from '@jupyterlab/apputils';

import { DocumentRegistry } from '@jupyterlab/docregistry';


/**
 * Initialization data for the teacher-ext extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'teacher-ext:plugin',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [IFileBrowserFactory],
  activate: (
      app: JupyterFrontEnd,
      nbTrack: INotebookTracker,
      browserFactory: IFileBrowserFactory | null,
      docManager: IDocumentManager,
      ) => {
    console.log('JupyterLab extension teacher-ext is activated!');

    nbTrack.currentChanged.connect(() => {

      const notebookPanel = nbTrack.currentWidget;
      const notebook = nbTrack.currentWidget.content;

      var cellContents : string[]=[];

      notebookPanel.context.ready.then(async () => {
       
        notebook.widgets.map(c => 
          cellContents.push(c.model.value.text)
        );

        let currentCell: Cell = null;
        let currentCellCheckButton: CellCheckButton = null;

        nbTrack.activeCellChanged.connect(() => {

          if (currentCell) {
            notebook.widgets.map((c: Cell) => {
              if (c.model.type == 'code') {

                const currentLayout = c.layout as PanelLayout;
                currentLayout.widgets.map(w => {
                  if (w === currentCellCheckButton) {
                    currentLayout.removeWidget(w)
                  }
                })

              }
              
            });
          }

          const cell: Cell = notebook.activeCell;

          const heading = cell.model.value.text.split("\n")[0].split(" ")

          var info : CellInfo = {
            id:  Number(heading[2]),
            question_id: Number(heading[1]),
            student_id: Number(heading[0].replace("#","")),
            code: cell.model.value.text.split("\n")[1],
          };


          const activeIndex = notebook.activeCellIndex

          if (activeIndex !== 0) {
            info.message = cellContents[activeIndex-1]
            info.feedback = cellContents[activeIndex+1]
          }

          const newCheckButton: CellCheckButton = new CellCheckButton(
            cell,info);

          (cell.layout as PanelLayout).addWidget(newCheckButton);

          // Set the current cell and button for future
          // reference
          currentCell = cell;
          currentCellCheckButton = newCheckButton;

        });

      });
    });

    //  tell the document registry about your widget extension:
    app.docRegistry.addWidgetExtension('Notebook', new ButtonExtension());
  
  }
};


export class ButtonExtension
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
    const getSubmissions = () => {
      NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;

      var item: CellInfo

      requestAPI<any>('code',{
        method: 'GET'
      })
        .then(data => {


          if (panel.context.path !== "Submissions/all_submissions.ipynb") {
            window.alert("Submissions Notebook not opened. \nGo to all_submissions.ipynb Notebook inside Submissions directory.")
            return
          }

          // Change Cell Type
          NotebookActions.changeCellType(notebook,'code')

          notebook.activeCellIndex = 0;
          for ( item of data.data) {

            // Insert message
            NotebookActions.insertBelow(notebook);
            notebook.activeCell.model.value.text = item.student_name + " @ " + item.time + " wrote: \n" + item.message;

            
            NotebookActions.changeCellType(notebook,'markdown')

            // Insert Code blocks:
            NotebookActions.insertBelow(notebook);
            notebook.activeCell.model.value.text = "#" + item.student_id + " " + item.question_id + " " + item.id + "\n" +  item.code;
           
            NotebookActions.changeCellType(notebook,'code')

            // Insert placeholder for Instructor feedback
            NotebookActions.insertBelow(notebook);
            notebook.activeCell.model.value.text = item.message;
            notebook.activeCell.model.value.text = "Instructor Feedback for " + item.student_name + ": \n" ;
            
            NotebookActions.changeCellType(notebook,'markdown')
                  
          }

        })
        .catch(reason => {
          window.alert("Failed to get recent submissions from server.\nPlease check your connection.")
          console.error(
            `Failed to get code from the server.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'sync-code-button',
      label: 'Get Submissions',
      onClick: getSubmissions,
      tooltip: 'Get available submissions from students.',
    });

    panel.toolbar.insertItem(10, 'getSubmissions', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export default plugin;
