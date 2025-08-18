import { ReactWidget } from '@jupyterlab/ui-components';

import { Cell, CodeCell } from '@jupyterlab/cells';

import { CellInfo } from './model';

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

const ShareButton = ({ icon, onClick }: IButtonComponent) => (
  <button type="button" onClick={() => onClick()} className="cellButton">
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
  info
}: ICodeCellButtonComponent): JSX.Element => {
  const shareCode = async () => {
    if (isNaN(info.problem_id)) {
      showErrorMessage(
        'Code Share Error',
        'Invalid code block. Use specific problem notebook.'
      );
      return;
    }

    const postBody = {
      message: info.message,
      code: cell.model.sharedModel.getSource(),
      problem_id: info.problem_id,
      snapshot: 2
    };
    console.log('From widget: ', postBody);
    requestAPI<any>('submissions', {
      method: 'POST',
      body: JSON.stringify(postBody)
    })
      .then(data => {
        if (data.msg === 'Submission saved successfully.') {
          data.msg = 'Code is sent to the instructor.';
        }
        showDialog({
          title: '',
          body: data.msg,
          buttons: [Dialog.okButton({ label: 'Ok' })]
        });
      })
      .catch(reason => {
        showErrorMessage('Code Share Error', reason);
        console.error(`Failed to share code to server.\n${reason}`);
      });
  };

  return (
    <div>
      <ShareButton icon={fileUploadIcon} onClick={() => shareCode()} />
    </div>
  );
};

export class FloatingFeedbackWidget {
  public node: HTMLDivElement;
  private container: HTMLElement;
  private isDragging = false;
  private dragOffset = { x: 0, y: 0 };
  private position = { x: 0, y: 50 }; // Will be calculated in setupContainer
  private size = { width: 300, height: 400 };
  private panelId: string;
  private filename: string;

  constructor(filename?: string) {
    this.panelId = filename || `feedback-${Date.now()}`;
    this.filename = filename || 'Unknown';
    this.node = document.createElement('div');
    this.setupContainer();
    this.setupEventListeners();
    this.createContent();
  }

  private setupContainer(): void {
    // Find the main content container similar to StickyLand
    this.container = document.querySelector('#jp-main-content-panel') as HTMLElement;
    if (!this.container) {
      this.container = document.querySelector('#main-panel') as HTMLElement;
    }
    if (!this.container) {
      this.container = document.body;
    }

    // Calculate top-right position
    const containerWidth = this.container.clientWidth || window.innerWidth;
    this.position.x = containerWidth - this.size.width - 20; // 20px margin from right edge
    this.position.y = 50; // 50px from top

    // Setup the floating window styles
    this.node.classList.add('floating-feedback-window');
    // Use filename as unique identifier, sanitize for valid DOM ID
    const sanitizedId = this.panelId.replace(/[^a-zA-Z0-9-_]/g, '-');
    this.node.id = `floating-feedback-${sanitizedId}`;
    this.node.style.position = 'absolute';
    this.node.style.left = `${this.position.x}px`;
    this.node.style.top = `${this.position.y}px`;
    this.node.style.width = `${this.size.width}px`;
    this.node.style.height = `${this.size.height}px`;
    this.node.style.zIndex = '1000';
    this.node.style.backgroundColor = '#ffffff';
    this.node.style.border = '1px solid #0078d4';
    this.node.style.borderRadius = '8px';
    this.node.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.15)';
    this.node.style.display = 'flex';
    this.node.style.flexDirection = 'column';
    this.node.style.overflow = 'hidden';
    this.node.style.fontFamily = 'var(--jp-ui-font-family)';
    this.node.style.fontSize = '13px';
  }

  private createContent(): void {
    // Create header
    const header = document.createElement('div');
    header.classList.add('feedback-header');
    header.style.padding = '12px';
    header.style.backgroundColor = '#0078d4';
    header.style.color = 'white';
    header.style.cursor = 'grab';
    header.style.userSelect = 'none';
    header.style.display = 'flex';
    header.style.justifyContent = 'space-between';
    header.style.alignItems = 'center';
    header.style.borderRadius = '8px 8px 0 0';

    const title = document.createElement('span');
    // Extract just the filename from the full path
    const displayName = this.filename.split('/').pop()?.replace('.ipynb', '') || 'Unknown';
    title.textContent = `📝 Feedback On ${displayName}`;
    title.style.fontWeight = '600';
    title.style.fontSize = '14px';

    const closeButton = document.createElement('button');
    closeButton.textContent = '×';
    closeButton.style.background = 'none';
    closeButton.style.border = 'none';
    closeButton.style.color = 'white';
    closeButton.style.fontSize = '18px';
    closeButton.style.cursor = 'pointer';
    closeButton.style.padding = '0';
    closeButton.style.width = '20px';
    closeButton.style.height = '20px';
    closeButton.style.borderRadius = '50%';
    closeButton.style.display = 'flex';
    closeButton.style.alignItems = 'center';
    closeButton.style.justifyContent = 'center';
    closeButton.addEventListener('click', () => this.close());
    closeButton.addEventListener('mouseenter', () => {
      closeButton.style.backgroundColor = 'rgba(255, 255, 255, 0.2)';
    });
    closeButton.addEventListener('mouseleave', () => {
      closeButton.style.backgroundColor = 'transparent';
    });

    header.appendChild(title);
    header.appendChild(closeButton);

    // Create content area
    const content = document.createElement('div');
    content.classList.add('feedback-content');
    content.style.flex = '1';
    content.style.padding = '16px';
    content.style.overflow = 'auto';
    content.style.backgroundColor = '#f8f9fa';

    const feedbackText = document.createElement('p');
    feedbackText.textContent = 'Your feedback will appear here...';
    feedbackText.style.margin = '0 0 12px 0';
    feedbackText.style.fontSize = '14px';
    feedbackText.style.color = '#333';

    content.appendChild(feedbackText);

    this.node.appendChild(header);
    this.node.appendChild(content);
  }

  private setupEventListeners(): void {
    this.node.addEventListener('mousedown', (e) => this.handleMouseDown(e));
    document.addEventListener('mousemove', (e) => this.handleMouseMove(e));
    document.addEventListener('mouseup', () => this.handleMouseUp());
  }

  private handleMouseDown(e: MouseEvent): void {
    const target = e.target as HTMLElement;
    if (target.classList.contains('feedback-header') || target.closest('.feedback-header')) {
      this.isDragging = true;
      const rect = this.node.getBoundingClientRect();
      this.dragOffset = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top
      };
      this.node.style.cursor = 'grabbing';
      const header = this.node.querySelector('.feedback-header') as HTMLElement;
      if (header) header.style.cursor = 'grabbing';
    }
  }

  private handleMouseMove(e: MouseEvent): void {
    if (!this.isDragging) return;

    const containerRect = this.container.getBoundingClientRect();
    const newX = Math.max(0, Math.min(
      this.container.clientWidth - this.size.width,
      e.clientX - containerRect.left - this.dragOffset.x
    ));
    const newY = Math.max(0, Math.min(
      this.container.clientHeight - this.size.height,
      e.clientY - containerRect.top - this.dragOffset.y
    ));

    this.position = { x: newX, y: newY };
    this.node.style.left = `${newX}px`;
    this.node.style.top = `${newY}px`;
  }

  private handleMouseUp(): void {
    if (this.isDragging) {
      this.isDragging = false;
      this.node.style.cursor = 'default';
      const header = this.node.querySelector('.feedback-header') as HTMLElement;
      if (header) header.style.cursor = 'grab';
    }
  }

  public show(): void {
    if (!this.node.parentElement) {
      this.container.appendChild(this.node);
      
      // Recalculate position in case container size changed
      const containerWidth = this.container.clientWidth || window.innerWidth;
      this.position.x = containerWidth - this.size.width - 20;
      this.node.style.left = `${this.position.x}px`;
    }
    this.node.style.display = 'flex';
  }

  public close(): void {
    if (this.node.parentElement) {
      this.node.parentElement.removeChild(this.node);
    }
  }

  public hide(): void {
    this.node.style.display = 'none';
  }
}

// Keep the ReactWidget for backward compatibility if needed
const FeedbackWidget = (): JSX.Element => {
  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#f8f9fa',
        fontFamily: 'var(--jp-ui-font-family)',
        fontSize: '13px'
      }}
    >
      <div 
        style={{
          padding: '12px',
          backgroundColor: '#0078d4',
          color: 'white',
          display: 'flex',
          alignItems: 'center',
          borderBottom: '1px solid #ccc'
        }}
      >
        <span style={{ fontWeight: '600', fontSize: '14px' }}>📝 Feedback</span>
      </div>
      <div style={{ 
        flex: 1, 
        padding: '16px', 
        overflow: 'auto',
        backgroundColor: '#ffffff'
      }}>
        <p style={{ margin: '0 0 12px 0', fontSize: '14px', color: '#333' }}>
          Your feedback will appear here...
        </p>
      </div>
    </div>
  );
};

export class FeedbackWidgetComponent extends ReactWidget {
  constructor() {
    super();
    this.addClass('jp-FeedbackWidget');
    this.id = 'carpo-feedback-widget';
  }
  render(): JSX.Element {
    return <FeedbackWidget />;
  }
}

export class CellCheckButton extends ReactWidget {
  cell: Cell = null;
  info: CellInfo = null;
  constructor(cell: Cell, info: CellInfo) {
    super();
    this.cell = cell;
    this.info = info;
    this.addClass('jp-CellButton');
  }
  render(): JSX.Element {
    return (
      <CodeCellButtonComponent
        cell={this.cell as CodeCell}
        info={this.info as CellInfo}
      />
    );
  }
}
