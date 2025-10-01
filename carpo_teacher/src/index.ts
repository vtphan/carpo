import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import { requestAPI } from './handler';

import {
  NotebookActions,
  NotebookPanel,
  INotebookModel,
  INotebookTracker

} from '@jupyterlab/notebook';

import { Cell } from '@jupyterlab/cells';

import { IFileBrowserFactory } from '@jupyterlab/filebrowser';

import {
  IDocumentManager
} from '@jupyterlab/docmanager';


import { IDisposable, DisposableDelegate } from '@lumino/disposable';
import { ToolbarButton, Dialog, showDialog, showErrorMessage } from '@jupyterlab/apputils';
import { Widget } from '@lumino/widgets';

// , InputDialog
import { DocumentRegistry } from '@jupyterlab/docregistry';

import { GetSolutionButton } from './upload-solution'

import { ICommandPalette } from '@jupyterlab/apputils';


const CommandIds = {
  /**
   * Command to run a code cell.
   */
  mainMenuRegister: 'jlab-carpo:main-register',
  mainMenuGotoApp: 'jlab-carpo:main-goto-app',
  mainMenuCollectNotebooks: 'jlab-carpo:main-download-notebooks',
  mainMenuAbout: 'jlab-carpo:main-about',

};

class RegistrationWidget extends Widget {
  private nameInput: HTMLInputElement;
  private serverUrlInput: HTMLInputElement;
  private appUrlInput: HTMLInputElement;

  constructor() {
    super();
    this.node.innerHTML = `
      <div style="padding: 20px; font-family: var(--jp-ui-font-family);">
        <div style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px; font-weight: 500;">Name:</label>
          <input type="text" id="name-input" style="width: 100%; padding: 8px; border: 1px solid var(--jp-border-color1); border-radius: 3px; font-size: 13px;" placeholder="Enter your name" />
        </div>
        <div style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px; font-weight: 500;">Server URL:</label>
          <input type="url" id="server-url-input" placeholder="http://127.0.0.1:8081" style="width: 100%; padding: 8px; border: 1px solid var(--jp-border-color1); border-radius: 3px; font-size: 13px;" placeholder="Enter server URL" />
        </div>
        <div style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px; font-weight: 500;">App URL:</label>
          <input type="url" id="app-url-input" placeholder="http://127.0.0.1:8080" style="width: 100%; padding: 8px; border: 1px solid var(--jp-border-color1); border-radius: 3px; font-size: 13px;" placeholder="Enter app URL" />
        </div>
      </div>
    `;

    this.nameInput = this.node.querySelector('#name-input') as HTMLInputElement;
    this.serverUrlInput = this.node.querySelector('#server-url-input') as HTMLInputElement;
    this.appUrlInput = this.node.querySelector('#app-url-input') as HTMLInputElement;
  }

  getValue() {
    return {
      name: this.nameInput.value,
      serverUrl: this.serverUrlInput.value,
      appUrl: this.appUrlInput.value
    };
  }
}

/**
 * Initialization data for the carpo-teacher extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'carpo-teacher:plugin',
  autoStart: true,
  requires: [INotebookTracker, ICommandPalette],
  optional: [IFileBrowserFactory],
  activate: (
      app: JupyterFrontEnd,
      nbTrack: INotebookTracker,
      palette: ICommandPalette,
      browserFactory: IFileBrowserFactory | null,
      docManager: IDocumentManager,
      ) => {
    console.log('JupyterLab extension carpo-teacher is activated!');

    const { commands } = app;

    const RegisterMenu = CommandIds.mainMenuRegister
    commands.addCommand(RegisterMenu, {
      label: 'Register',
      caption: 'Register user to server.',
      execute: async (args: any) => {
        try {
          const registrationWidget = new RegistrationWidget();
          
          const result = await showDialog({
            title: 'Registration Information',
            body: registrationWidget,
            buttons: [Dialog.cancelButton(), Dialog.okButton({ label: 'Register' })]
          });
          
          if (!result.button.accept) {
            return;
          }
          
          const formData = registrationWidget.getValue();
          
          // Validate that all fields are filled
          if (!formData.name || !formData.serverUrl || !formData.appUrl) {
            showErrorMessage('Registration Error', 'Please fill in all required fields.');
            return;
          }
          
          // Send POST request with collected information
          requestAPI<any>('register', {
            method: 'POST',
            body: JSON.stringify(formData)
          })
          .then(data => {
            console.log('Registration successful:', data);
            showDialog({
              title: 'Registration Successful',
              body: `User ${formData.name} has been registered successfully.`,
              buttons: [Dialog.okButton({ label: 'Ok' })]
            });
          })
          .catch(reason => {
            showErrorMessage('Registration Error', reason);
            console.error(`Failed to register user.\n${reason}`);
          });
          
        } catch (error) {
          console.error('Registration dialog error:', error);
          showErrorMessage('Registration Error', 'Failed to collect registration information.');
        }
      }
    });

    // Add the command to the command palette
    const category = 'Extension Examples';
    palette.addItem({
      command: RegisterMenu,
      category: category,
      args: { origin: 'from the palette' }
    });

    const GotoAppMenu = CommandIds.mainMenuGotoApp
    commands.addCommand(GotoAppMenu, {
      label: 'Go to App',
      caption: 'Open the web app.',
      execute: (args: any) => {
        console.log("Args: ", args)
        requestAPI<any>('view_app',{
          method: 'GET'
        })
          .then(data => {
            console.log(data);
            window.open(
              data.url, "_blank");
          })
          .catch(reason => {
            showErrorMessage('View App Status Error', reason);
            console.error(
              `Failed to view app status.\n${reason}`
            );
          });
      }
    });

    // Add the command to the command palette
    palette.addItem({
      command: GotoAppMenu,
      category: category,
      args: { origin: 'from the palette' }
    });

    const DownloadNotebooks = CommandIds.mainMenuCollectNotebooks
    commands.addCommand(DownloadNotebooks, {
      label: 'Download Notebooks',
      caption: 'Download all students Notebooks.',
      execute: async (args: any) => {
        console.log("Args: ", args)
        const widget = new NotebookSelectWidget();

        const result = await showDialog({
          title: 'Select Notebook Name',
          body: widget,
          buttons: [Dialog.cancelButton(), Dialog.okButton({ label: 'Download' })]
        });
        
        if (!result.button.accept) {
          return;
        }

        const selectedNotebookId = widget.value;
        const notebookName = widget.name
        console.log('Selected notebook ID:', selectedNotebookId, notebookName);
        console.log('selected: ', widget)

        const data = await requestAPI<any>(`notebooks/${selectedNotebookId}/downloads?name=${notebookName}`, {
          method: 'GET'
        });
        showDialog({
          title: 'Downloads',
          body: data.message,
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });
      }
    });

    // Add the command to the command palette
    palette.addItem({
      command: DownloadNotebooks,
      category: category,
      args: { origin: 'from the palette' }
    });

    const AboutMenu = CommandIds.mainMenuAbout
    commands.addCommand(AboutMenu, {
      label: 'About Carpo',
      caption: 'Carpo Information',
      execute: (args: any) => {
        const content = new Widget();
        content.node.innerHTML = `
          <h3>How to use carpo:</h3>
          <ol>
            <li><strong>To Register </strong>: Input name, ServerUrl and AppUrl. </li>
            <li><strong>Publish</strong>: To publish current cell as an exercise.</li>
            <li><strong>Unpublish</strong>: To publish an exercise.</li>
            <li><strong>UploadSolution</strong>: To upload the exercise solution.</li>
          </ol>
        `;

        showDialog({
          title: 'About Carpo',
          body: content,
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });
      }
    });

    // Add the command to the command palette
    palette.addItem({
      command: AboutMenu,
      category: category,
      args: { origin: 'from the palette' }
    });

    //  tell the document registry about your widget extension:
    // app.docRegistry.addWidgetExtension('Notebook', new RegisterButton());
    // app.docRegistry.addWidgetExtension('Notebook', new GoToApp());
    app.docRegistry.addWidgetExtension('Notebook', new PublishProblemButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new ArchiveProblemButtonExtension());
    app.docRegistry.addWidgetExtension('Notebook', new GetSolutionButton());

  }
};

// deprecated
export class RegisterButton
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
    const register = () => {

      requestAPI<any>('register',{
        method: 'GET'
      })
        .then(data => {
          console.log(data);

          showDialog({
            title:'',
            body:  "Instructor "+ data.name + " has been registered.",
            buttons: [Dialog.okButton({ label: 'Ok' })]
          });
         
        })
        .catch(reason => {
          showErrorMessage('Registration Error', reason);
          console.error(
            `Failed to register user as Instructor.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'register-button',
      label: 'Register',
      onClick: register,
      tooltip: 'Register as a Teacher',
    });

    panel.toolbar.insertItem(10, 'register', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

// deprecated
export class GoToApp implements DocumentRegistry.IWidgetExtension<NotebookPanel, INotebookModel>
{
  createNew(
    panel: NotebookPanel,
    context: DocumentRegistry.IContext<INotebookModel>
  ): IDisposable {
    const viewWebApp = () => {

    requestAPI<any>('view_app',{
      method: 'GET'
    })
      .then(data => {
        // console.log(data);
        window.open(
          data.url, "_blank");
      })
      .catch(reason => {
        showErrorMessage('View App Status Error', reason);
        console.error(
          `Failed to view app status.\n${reason}`
        );
      });

    };

    const button = new ToolbarButton({
      className: 'get-app-button',
      label: 'App',
      onClick: viewWebApp,
      tooltip: 'Go to the web app',
    });

    panel.toolbar.insertItem(11, 'viewWebApp', button);

    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class PublishProblemButtonExtension
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
    const publishProblem = () => {
      NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;
      const activeIndex = notebook.activeCellIndex
      var problem:string
      var format:string
      var header:string
      var time_limit:string

      notebook.widgets.map((c:Cell, index:number) => {
        if (index === activeIndex ) {
          problem = c.model.sharedModel.getSource()
          format = c.model.type
        }
      });

      if (problem.includes("#PID:")) {
        showErrorMessage('Publish Question Error', "Problem already published.")
        return
      }

      if (!problem) {
        showErrorMessage('Publish Question Error', "Problem is empty.")
        return
      }


      header = problem.split('\n')[0]
      if(header.match(/[0-9]+[a-zA-Z]/)) {
        time_limit = header.match(/[0-9]+[a-zA-Z]/)[0]
      }

      let postBody = {
        "question": problem,
        "format": format,
        "time_limit": time_limit
      }

      requestAPI<any>('problem',{
        method: 'POST',
        body: JSON.stringify(postBody)

      })
        .then(data => {
          // console.log(data)
          notebook.widgets.map((c:Cell,index:number) => {
            if (index === activeIndex ) {
              c.model.sharedModel.setSource("#PID:" + data.id + "\n" + problem)
            }
          });


          showDialog({
          title:'New Questions Published',
          body: 'Problem ' + data.id + " is published.",
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });

        })
        .catch(reason => {
          showErrorMessage('Publish Question Error', reason);
          console.error(
            `Failed to publish question to the server.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'publish-problem-button',
      label: 'Publish',
      onClick: publishProblem,
      tooltip: 'Publish New Problem.',
    });

    panel.toolbar.insertItem(10, 'publishNewProblem', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

export class ArchiveProblemButtonExtension
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
    const archiveProblem = () => {
      NotebookActions.clearAllOutputs(panel.content);

      const notebook = panel.content;
      const activeIndex = notebook.activeCellIndex
      var problem:string

      notebook.widgets.map((c:Cell,index:number) => {
        if (index === activeIndex ) {
          problem = c.model.sharedModel.getSource()
        }
      });

      if (!problem.includes("#PID:")) {
        showErrorMessage('Unpublish Question Error', "Active problem not found.")
        return
      }

      var problem_id: number = parseInt((problem.split("\n")[0]).split("#PID:")[1]);

      let body = {
        "problem_id": problem_id
      }

      requestAPI<any>('problem',{
        method: 'DELETE',
        body: JSON.stringify(body)

      })
        .then(data => {
          console.log(data)
          showDialog({
          title:'Question Unpublished',
          body: 'Problem id ' + problem_id +' is  unpublished.',
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });

        })
        .catch(reason => {
          showErrorMessage('Unpublish Question Error', reason);
          console.error(
            `Failed to unpublish question.\n${reason}`
          );
        });

    };

    const button = new ToolbarButton({
      className: 'archive-problem-button',
      label: 'Unpublish',
      onClick: archiveProblem,
      tooltip: 'Unpublish the problem.',
    });

    panel.toolbar.insertItem(11, 'archivesProblem', button);
    return new DisposableDelegate(() => {
      button.dispose();
    });
  }
}

class NotebookSelectWidget extends Widget {
  private select: HTMLSelectElement;

  constructor() {
    super();
    this.node.classList.add('my-dropdown-widget');
    this.select = document.createElement('select');
    this.select.innerHTML = `<option value="">Loading...</option>`;
    
    this.loadNotebooks();

    this.node.appendChild(this.select);
  }

  private async loadNotebooks() {
    try {
      const data = await requestAPI<any>('notebooks', {
        method: 'GET'
      });
      
      this.select.innerHTML = '';
      console.log(data)
      
      if (data && data.data && data.data.length > 0) {
        data.data.forEach((notebook: any) => {
          const option = document.createElement('option');
          option.value = notebook.id;
          option.text = notebook.title || `Notebook ${notebook.id}`;
          this.select.appendChild(option);
        });
      } else {
        this.select.innerHTML = '<option value="">No notebooks available</option>';
      }
      
    } catch (error) {
      console.error('Failed to load notebooks:', error);
      this.select.innerHTML = '<option value="">Error loading notebooks</option>';
    }
  }

  get value(): string {
    return this.select.value;
  }

  get name(): string {
    return this.select.options[this.select.selectedIndex]?.text || '';
  }
}



export default plugin;
