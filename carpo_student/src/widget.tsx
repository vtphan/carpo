import { ReactWidget } from '@jupyterlab/ui-components';

import { Cell, CodeCell } from '@jupyterlab/cells';

import { CellInfo } from './model';

import { fileUploadIcon, LabIcon } from '@jupyterlab/ui-components';

import React from 'react';
import { requestAPI } from './handler';
import { Dialog, showDialog, showErrorMessage } from '@jupyterlab/apputils';
import { initializeNotifications } from './sse-notifications';

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
    // console.log('From widget: ', postBody);
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

      initializeNotifications()
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
  private isResizing = false;
  private resizeCorner: 'nw' | 'ne' | 'sw' | 'se' | null = null;
  private dragOffset = { x: 0, y: 0 };
  private resizeOffset = { x: 0, y: 0 };
  private position = { x: 0, y: 50 }; // Will be calculated in setupContainer
  private size = { width: 550, height: 400 };
  private minSize = { width: 250, height: 200 };
  private panelId: string;
  private filename: string;
  private contentElement: HTMLDivElement;
  private isLoading = false;

  constructor(filename?: string) {
    this.panelId = filename || `feedback-${Date.now()}`;
    this.filename = filename || 'Unknown';
    this.node = document.createElement('div');
    this.setupContainer();
    this.setupEventListeners();
    this.createContent();
    this.fetchFeedbackContent();
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

    // Calculate bottom-right position
    const containerWidth = this.container.clientWidth || window.innerWidth;
    const containerHeight = this.container.clientHeight || window.innerHeight;
    this.position.x = containerWidth - this.size.width - 20; // 20px margin from right edge
    this.position.y = containerHeight - this.size.height - 40;; // 20px from bottom

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
    this.node.style.resize = 'none'; // Disable browser default resize
    
    // Add resize handle
    this.createResizeHandle();
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

    // Create button container for refresh and close buttons
    const buttonContainer = document.createElement('div');
    buttonContainer.style.display = 'flex';
    buttonContainer.style.gap = '4px';

    // Refresh button
    const refreshButton = document.createElement('button');
    refreshButton.textContent = '🔄';
    refreshButton.style.background = 'none';
    refreshButton.style.border = 'none';
    refreshButton.style.color = 'white';
    refreshButton.style.fontSize = '14px';
    refreshButton.style.cursor = 'pointer';
    refreshButton.style.padding = '0';
    refreshButton.style.width = '20px';
    refreshButton.style.height = '20px';
    refreshButton.style.borderRadius = '50%';
    refreshButton.style.display = 'flex';
    refreshButton.style.alignItems = 'center';
    refreshButton.style.justifyContent = 'center';
    refreshButton.title = 'Refresh feedback';
    refreshButton.addEventListener('click', () => this.refreshFeedback());
    refreshButton.addEventListener('mouseenter', () => {
      refreshButton.style.backgroundColor = 'rgba(255, 255, 255, 0.2)';
    });
    refreshButton.addEventListener('mouseleave', () => {
      refreshButton.style.backgroundColor = 'transparent';
    });

    // Close button
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
    closeButton.title = 'Close feedback';
    closeButton.addEventListener('click', () => this.close());
    closeButton.addEventListener('mouseenter', () => {
      closeButton.style.backgroundColor = 'rgba(255, 255, 255, 0.2)';
    });
    closeButton.addEventListener('mouseleave', () => {
      closeButton.style.backgroundColor = 'transparent';
    });

    buttonContainer.appendChild(refreshButton);
    buttonContainer.appendChild(closeButton);

    header.appendChild(title);
    header.appendChild(buttonContainer);

    // Create content area
    const content = document.createElement('div');
    content.classList.add('feedback-content');
    content.style.flex = '1';
    content.style.padding = '4px';
    content.style.overflow = 'auto';
    content.style.backgroundColor = '#ffffff';

    this.contentElement = content;

    this.node.appendChild(header);
    this.node.appendChild(content);
  }

  private createResizeHandle(): void {
    const corners = [
      { corner: 'nw', top: '0', left: '0', cursor: 'nw-resize', title: 'Resize from top-left' },
      { corner: 'ne', top: '0', right: '0', cursor: 'ne-resize', title: 'Resize from top-right' },
      { corner: 'sw', bottom: '0', left: '0', cursor: 'sw-resize', title: 'Resize from bottom-left' },
      { corner: 'se', bottom: '0', right: '0', cursor: 'se-resize', title: 'Resize from bottom-right' }
    ];

    corners.forEach(({ corner, cursor, title, ...position }) => {
      const resizeHandle = document.createElement('div');
      resizeHandle.classList.add('feedback-resize-handle', `resize-${corner}`);
      resizeHandle.dataset.corner = corner;
      resizeHandle.style.position = 'absolute';
      resizeHandle.style.width = '15px';
      resizeHandle.style.height = '15px';
      resizeHandle.style.cursor = cursor;
      resizeHandle.style.opacity = '0.7';
      resizeHandle.style.zIndex = '10';
      resizeHandle.title = title;

      // Set position based on corner
      Object.entries(position).forEach(([key, value]) => {
        if (value !== undefined) {
          (resizeHandle.style as any)[key] = value;
        }
      });

      // Corner-specific styling
      const gradientMap = {
        'nw': 'linear-gradient(315deg, transparent 30%, #0078d4 30%, #0078d4 60%, transparent 60%)',
        'ne': 'linear-gradient(225deg, transparent 30%, #0078d4 30%, #0078d4 60%, transparent 60%)',
        'sw': 'linear-gradient(45deg, transparent 30%, #0078d4 30%, #0078d4 60%, transparent 60%)',
        'se': 'linear-gradient(135deg, transparent 30%, #0078d4 30%, #0078d4 60%, transparent 60%)'
      };
      
      resizeHandle.style.background = gradientMap[corner as keyof typeof gradientMap];
      resizeHandle.style.backgroundSize = '8px 8px';
      
      // Add hover effect
      resizeHandle.addEventListener('mouseenter', () => {
        resizeHandle.style.opacity = '1';
      });
      resizeHandle.addEventListener('mouseleave', () => {
        resizeHandle.style.opacity = '0.7';
      });

      this.node.appendChild(resizeHandle);
    });
  }

  private async fetchFeedbackContent(): Promise<void> {
    this.isLoading = true;
    this.showLoadingState();

    try {
      // Extract problem ID from filename (e.g., "ex001.ipynb" -> "1")
      const problemId = this.extractProblemId(this.filename);
      
      const resp = await requestAPI<any>(`widget-feedback?problem_id=${problemId}`, {
        method: 'GET'
      });
      
      this.showFeedbackContent(resp.data || 'No feedback available');
    } catch (error) {
      console.error('Failed to fetch feedback:', error);
      this.showErrorState(error);
    } finally {
      this.isLoading = false;
    }
  }

  private extractProblemId(filename: string): number {
    // Extract filename from path (e.g., "Exercises/ex001.ipynb" -> "ex001.ipynb")
    const basename = filename.split('/').pop() || '';
    
    // Extract number from filename (e.g., "ex001.ipynb" -> "001")
    const match = basename.match(/ex(\d+)\.ipynb$/);
    if (match) {
      return parseInt(match[1], 10); // Convert "001" to 1
    }
    
    // Fallback to 1 if no match found
    return 1;
  }

  private showLoadingState(): void {
    this.contentElement.innerHTML = '';
    const loadingDiv = document.createElement('div');
    loadingDiv.style.textAlign = 'center';
    loadingDiv.style.color = '#666';
    loadingDiv.style.fontSize = '13px';
    loadingDiv.style.padding = '20px';
    loadingDiv.innerHTML = '🔄 Loading feedback...';
    this.contentElement.appendChild(loadingDiv);
  }

  private showFeedbackContent(data: any): void {
    this.contentElement.innerHTML = '';
    
    // Handle array of feedback objects
    if (Array.isArray(data)) {
      this.createChatMessages(data);
    } else {
      // Handle single feedback object or no data
      const feedbackDiv = document.createElement('div');
      feedbackDiv.style.fontSize = '14px';
      feedbackDiv.style.color = '#666';
      feedbackDiv.style.textAlign = 'center';
      feedbackDiv.style.padding = '20px';
      feedbackDiv.textContent = 'No feedback available';
      this.contentElement.appendChild(feedbackDiv);
    }
  }

  private createChatMessages(feedbackArray: any[]): void {
    const chatContainer = document.createElement('div');
    chatContainer.style.display = 'flex';
    chatContainer.style.flexDirection = 'column';
    chatContainer.style.gap = '12px';
    chatContainer.style.padding = '8px';
    
    feedbackArray.forEach((feedback, index) => {
      const messageDiv = this.createChatMessage(feedback, index);
      chatContainer.appendChild(messageDiv);
    });
    
    this.contentElement.appendChild(chatContainer);
  }

  private createChatMessage(feedback: any, index: number): HTMLDivElement {
    const messageContainer = document.createElement('div');
    messageContainer.style.display = 'flex';
    messageContainer.style.flexDirection = 'column';
    messageContainer.style.marginBottom = '8px';
    
    // Message bubble
    const messageBubble = document.createElement('div');
    messageBubble.style.backgroundColor = '#e3f2fd';
    messageBubble.style.border = '1px solid #2196f3';
    messageBubble.style.borderRadius = '12px';
    messageBubble.style.padding = '12px';
    messageBubble.style.fontSize = '13px';
    messageBubble.style.lineHeight = '1.4';
    messageBubble.style.wordWrap = 'break-word';
    messageBubble.style.maxWidth = '100%';
    messageBubble.style.position = 'relative';

    // Feedback content
    if ((feedback.has_feedback && feedback.code) || feedback.score) {
      const feedbackContent = document.createElement('pre');
      feedbackContent.style.marginBottom = '12px';
      feedbackContent.style.fontSize = '13px';
      feedbackContent.style.lineHeight = '1.4';
      feedbackContent.style.color = '#333';
      feedbackContent.style.whiteSpace = "pre-wrap";
      
      feedbackContent.innerHTML = feedback.code;
      if (feedback.score === 1 || feedback.score === 2) 
      { 
        feedbackContent.innerHTML += feedback.score === 1 ? "\n Your submission was graded correct.": "\n Your submission was graded incorrect.";
      }
      messageBubble.appendChild(feedbackContent);
    } 
    
    // Footer with star rating and timestamp
    const footer = document.createElement('div');
    footer.style.display = 'flex';
    footer.style.justifyContent = 'space-between';
    footer.style.alignItems = 'center';
    footer.style.marginTop = '8px';
    footer.style.paddingTop = '4px';
    footer.style.borderTop = '1px solid #eee';

    // Upvote/Downvote on the left
    const votingButtons = this.createVotingButtons(feedback.id || `${index}`);
    footer.appendChild(votingButtons);

    // Timestamp on the right
    const timestamp = document.createElement('div');
    timestamp.style.fontSize = '11px';
    timestamp.style.color = '#666';
    timestamp.style.opacity = '0.7';
    
    if (feedback.created_at) {
      // Format timestamp nicely
      const date = new Date(feedback.created_at);
      const timeString = date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
      timestamp.textContent = timeString;
    } else {
      timestamp.textContent = `Message ${index + 1}`;
    }
    
    footer.appendChild(timestamp);
    messageBubble.appendChild(footer);
    messageContainer.appendChild(messageBubble);
    return messageContainer;
  }

  private showErrorState(error: any): void {
    this.contentElement.innerHTML = '';
    const errorDiv = document.createElement('div');
    errorDiv.style.color = '#d73a49';
    errorDiv.style.fontSize = '13px';
    errorDiv.style.padding = '20px';
    errorDiv.style.textAlign = 'center';
    
    const errorMessage = error?.message || error || 'Failed to load feedback';
    errorDiv.innerHTML = `⚠️ Error loading feedback<br><small>${errorMessage}</small>`;
    
    const retryButton = document.createElement('button');
    retryButton.textContent = '🔄 Retry';
    retryButton.style.marginTop = '10px';
    retryButton.style.padding = '5px 10px';
    retryButton.style.backgroundColor = '#0078d4';
    retryButton.style.color = 'white';
    retryButton.style.border = 'none';
    retryButton.style.borderRadius = '4px';
    retryButton.style.cursor = 'pointer';
    retryButton.style.fontSize = '12px';
    retryButton.addEventListener('click', () => this.fetchFeedbackContent());
    
    errorDiv.appendChild(document.createElement('br'));
    errorDiv.appendChild(retryButton);
    this.contentElement.appendChild(errorDiv);
  }

  private setupEventListeners(): void {
    this.node.addEventListener('mousedown', (e) => this.handleMouseDown(e));
    document.addEventListener('mousemove', (e) => this.handleMouseMove(e));
    document.addEventListener('mouseup', () => this.handleMouseUp());
  }

  private handleMouseDown(e: MouseEvent): void {
    const target = e.target as HTMLElement;
    
    // Check for resize handle
    if (target.classList.contains('feedback-resize-handle')) {
      this.isResizing = true;
      this.resizeCorner = target.dataset.corner as 'nw' | 'ne' | 'sw' | 'se';
      const rect = this.node.getBoundingClientRect();
      
      // Set resize offset based on corner
      switch (this.resizeCorner) {
        case 'nw':
          this.resizeOffset = { x: e.clientX - rect.left, y: e.clientY - rect.top };
          break;
        case 'ne':
          this.resizeOffset = { x: e.clientX - (rect.left + rect.width), y: e.clientY - rect.top };
          break;
        case 'sw':
          this.resizeOffset = { x: e.clientX - rect.left, y: e.clientY - (rect.top + rect.height) };
          break;
        case 'se':
          this.resizeOffset = { x: e.clientX - (rect.left + rect.width), y: e.clientY - (rect.top + rect.height) };
          break;
      }
      
      e.preventDefault();
      e.stopPropagation();
      return;
    }
    
    // Check for header (drag functionality)
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
    if (this.isResizing && this.resizeCorner) {
      const containerRect = this.container.getBoundingClientRect();
      const mouseX = e.clientX - containerRect.left;
      const mouseY = e.clientY - containerRect.top;
      
      let newX = this.position.x;
      let newY = this.position.y;
      let newWidth = this.size.width;
      let newHeight = this.size.height;

      switch (this.resizeCorner) {
        case 'nw':
          // Top-left corner: resize from top-left
          newWidth = Math.max(this.minSize.width, this.position.x + this.size.width - (mouseX - this.resizeOffset.x));
          newHeight = Math.max(this.minSize.height, this.position.y + this.size.height - (mouseY - this.resizeOffset.y));
          newX = Math.min(this.position.x + this.size.width - this.minSize.width, mouseX - this.resizeOffset.x);
          newY = Math.min(this.position.y + this.size.height - this.minSize.height, mouseY - this.resizeOffset.y);
          break;
          
        case 'ne':
          // Top-right corner: resize from top-right
          newWidth = Math.max(this.minSize.width, mouseX - this.position.x - this.resizeOffset.x);
          newHeight = Math.max(this.minSize.height, this.position.y + this.size.height - (mouseY - this.resizeOffset.y));
          newY = Math.min(this.position.y + this.size.height - this.minSize.height, mouseY - this.resizeOffset.y);
          break;
          
        case 'sw':
          // Bottom-left corner: resize from bottom-left
          newWidth = Math.max(this.minSize.width, this.position.x + this.size.width - (mouseX - this.resizeOffset.x));
          newHeight = Math.max(this.minSize.height, mouseY - this.position.y - this.resizeOffset.y);
          newX = Math.min(this.position.x + this.size.width - this.minSize.width, mouseX - this.resizeOffset.x);
          break;
          
        case 'se':
          // Bottom-right corner: resize from bottom-right
          newWidth = Math.max(this.minSize.width, mouseX - this.position.x - this.resizeOffset.x);
          newHeight = Math.max(this.minSize.height, mouseY - this.position.y - this.resizeOffset.y);
          break;
      }

      // Apply boundary constraints
      newX = Math.max(0, Math.min(newX, this.container.clientWidth - newWidth));
      newY = Math.max(0, Math.min(newY, this.container.clientHeight - newHeight));
      newWidth = Math.min(newWidth, this.container.clientWidth - newX);
      newHeight = Math.min(newHeight, this.container.clientHeight - newY);

      this.position = { x: newX, y: newY };
      this.size = { width: newWidth, height: newHeight };
      this.node.style.left = `${newX}px`;
      this.node.style.top = `${newY}px`;
      this.node.style.width = `${newWidth}px`;
      this.node.style.height = `${newHeight}px`;
      return;
    }

    if (this.isDragging) {
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
  }

  private handleMouseUp(): void {
    if (this.isDragging) {
      this.isDragging = false;
      this.node.style.cursor = 'default';
      const header = this.node.querySelector('.feedback-header') as HTMLElement;
      if (header) header.style.cursor = 'grab';
    }
    
    if (this.isResizing) {
      this.isResizing = false;
      this.resizeCorner = null;
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
    
    // Refresh feedback content when shown
    if (!this.isLoading) {
      this.fetchFeedbackContent();
    }
  }

  public close(): void {
    if (this.node.parentElement) {
      this.node.parentElement.removeChild(this.node);
    }
  }

  public hide(): void {
    this.node.style.display = 'none';
  }

  public refreshFeedback(): void {
    if (!this.isLoading) {
      this.fetchFeedbackContent();
    }
  }

  private createVotingButtons(feedbackId: string): HTMLElement {
    const voteContainer = document.createElement('div');
    voteContainer.style.display = 'flex';
    voteContainer.style.gap = '8px';
    voteContainer.style.alignItems = 'center';
    voteContainer.dataset.feedbackId = feedbackId;

    // Create upvote button
    const upvoteBtn = document.createElement('button');
    upvoteBtn.textContent = '👍';
    upvoteBtn.style.fontSize = '16px';
    upvoteBtn.style.cursor = 'pointer';
    upvoteBtn.style.border = 'none';
    upvoteBtn.style.background = 'none';
    upvoteBtn.style.padding = '2px 4px';
    upvoteBtn.style.borderRadius = '4px';
    upvoteBtn.style.transition = 'background-color 0.2s ease';
    upvoteBtn.dataset.vote = '1';

    // Create downvote button
    const downvoteBtn = document.createElement('button');
    downvoteBtn.textContent = '👎';
    downvoteBtn.style.fontSize = '16px';
    downvoteBtn.style.cursor = 'pointer';
    downvoteBtn.style.border = 'none';
    downvoteBtn.style.background = 'none';
    downvoteBtn.style.padding = '2px 4px';
    downvoteBtn.style.borderRadius = '4px';
    downvoteBtn.style.transition = 'background-color 0.2s ease';
    downvoteBtn.dataset.vote = '-1';

    // Add hover effects
    upvoteBtn.addEventListener('mouseenter', () => {
      upvoteBtn.style.backgroundColor = 'rgba(34, 197, 94, 0.2)'; // Light green
    });
    upvoteBtn.addEventListener('mouseleave', () => {
      if (voteContainer.dataset.currentVote !== '1') {
        upvoteBtn.style.backgroundColor = 'transparent';
      }
    });

    downvoteBtn.addEventListener('mouseenter', () => {
      downvoteBtn.style.backgroundColor = 'rgba(239, 68, 68, 0.2)'; // Light red
    });
    downvoteBtn.addEventListener('mouseleave', () => {
      if (voteContainer.dataset.currentVote !== '-1') {
        downvoteBtn.style.backgroundColor = 'transparent';
      }
    });

    // Add click handlers
    upvoteBtn.addEventListener('click', () => {
      this.setVote(voteContainer, 1, feedbackId);
    });

    downvoteBtn.addEventListener('click', () => {
      this.setVote(voteContainer, -1, feedbackId);
    });

    voteContainer.appendChild(upvoteBtn);
    voteContainer.appendChild(downvoteBtn);

    // Load existing vote if any
    this.loadExistingVote(voteContainer, feedbackId);

    return voteContainer;
  }

  private updateVoteButtons(container: HTMLElement, vote: number): void {
    const upvoteBtn = container.querySelector('[data-vote="1"]') as HTMLElement;
    const downvoteBtn = container.querySelector('[data-vote="-1"]') as HTMLElement;
    
    // Reset both buttons
    upvoteBtn.style.backgroundColor = 'transparent';
    downvoteBtn.style.backgroundColor = 'transparent';
    
    // Highlight the selected vote
    if (vote === 1) {
      upvoteBtn.style.backgroundColor = 'rgba(34, 197, 94, 0.3)'; // Green for upvote
    } else if (vote === -1) {
      downvoteBtn.style.backgroundColor = 'rgba(239, 68, 68, 0.3)'; // Red for downvote
    }
  }

  private setVote(container: HTMLElement, vote: number, feedbackId: string): void {
    const currentVote = parseInt(container.dataset.currentVote || '0');
    
    // If clicking the same vote, remove it (toggle off)
    if (currentVote === vote) {
      container.dataset.currentVote = '0';
      this.updateVoteButtons(container, 0);
      localStorage.removeItem(`feedback_vote_${feedbackId}`);
      console.log(`Removed vote for feedback ${feedbackId}`);
      
      // Save removal to server (vote = 0)
      this.saveVoteToServer(feedbackId, 0);
    } else {
      // Set new vote
      container.dataset.currentVote = vote.toString();
      this.updateVoteButtons(container, vote);
      
      // Store vote in localStorage for persistence
      const storageKey = `feedback_vote_${feedbackId}`;
      localStorage.setItem(storageKey, vote.toString());
      
      console.log(`${vote === 1 ? 'Upvoted' : 'Downvoted'} feedback ${feedbackId}`);
      
      // Save vote to server
      this.saveVoteToServer(feedbackId, vote);
    }
  }

  private loadExistingVote(container: HTMLElement, feedbackId: string): void {
    const storageKey = `feedback_vote_${feedbackId}`;
    const existingVote = localStorage.getItem(storageKey);
    
    if (existingVote) {
      const vote = parseInt(existingVote);
      container.dataset.currentVote = vote.toString();
      this.updateVoteButtons(container, vote);
    }
  }

  private async saveVoteToServer(feedbackId: string, vote: number): Promise<void> {
    try {
      const response = await requestAPI<any>(`feedback-ratings`, {
        method: 'PUT',
        body: JSON.stringify({ id: feedbackId, rating: vote })
      });
      console.log('Vote saved to server:', response);
    } catch (error) {
      console.error('Failed to save vote to server:', error);
      // Could show a toast notification here if needed
    }
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
