// import { URLExt } from '@jupyterlab/coreutils';
// import { ServerConnection } from '@jupyterlab/services';
import { FloatingFeedbackWidget } from './widget';
import { requestAPI } from './handler';

interface Config {
  server: string;
  id: string;
}

async function fetchConfig(): Promise<Config> {
  try {
    const config = await requestAPI<any>('config', {
      method: 'GET'
    });
    return {
      server: config.server,
      id: config.id
    };
  } catch (error) {
    console.error('Failed to fetch config from API:', error);
    // Fallback values
    return {
      server: 'http://127.0.0.1:8081',
      id: '1'
    };
  }
}

export interface NotificationData {
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
  timestamp?: string;
  title?: string;
  filename?: string;
}

export class ToastNotification {
  private static container: HTMLDivElement | null = null;
  private static toastCounter = 0;
  static feedbackWidgets = new Map<string, FloatingFeedbackWidget>();

  static init(): void {
    if (!this.container) {
      this.container = document.createElement('div');
      this.container.id = 'toast-notification-container';
      this.container.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 10000;
        pointer-events: none;
        max-width: 400px;
      `;
      document.body.appendChild(this.container);
    }
  }

  static show(data: NotificationData): void {
    this.init();
    
    const toast = document.createElement('div');
    const toastId = `toast-${++this.toastCounter}`;
    toast.id = toastId;
    toast.style.cssText = `
      background: ${this.getBackgroundColor(data.type)};
      color: white;
      padding: 12px 16px;
      margin-bottom: 8px;
      border-radius: 6px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      pointer-events: auto;
      cursor: pointer;
      font-family: var(--jp-ui-font-family, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif);
      font-size: 13px;
      line-height: 1.4;
      min-width: 250px;
      max-width: 400px;
      word-wrap: break-word;
      opacity: 0;
      transform: translateX(100%);
      transition: all 0.3s ease-in-out;
      position: relative;
      border-left: 4px solid ${this.getBorderColor(data.type)};
    `;

    // Create toast content
    const content = document.createElement('div');
    if (data.title) {
      const title = document.createElement('div');
      title.style.cssText = `
        font-weight: 600;
        margin-bottom: 4px;
        font-size: 14px;
      `;
      title.textContent = data.title;
      content.appendChild(title);
    }

    const message = document.createElement('div');
    message.textContent = data.message;
    content.appendChild(message);

    // Add timestamp if provided
    if (data.timestamp) {
      const time = document.createElement('div');
      time.style.cssText = `
        font-size: 11px;
        opacity: 0.8;
        margin-top: 4px;
      `;
      time.textContent = new Date(data.timestamp).toLocaleTimeString();
      content.appendChild(time);
    }

    // Add close button
    const closeBtn = document.createElement('button');
    closeBtn.innerHTML = '×';
    closeBtn.style.cssText = `
      position: absolute;
      top: 8px;
      right: 8px;
      background: none;
      border: none;
      color: white;
      font-size: 16px;
      cursor: pointer;
      padding: 0;
      width: 20px;
      height: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      opacity: 0.7;
      transition: opacity 0.2s;
    `;
    closeBtn.onmouseover = () => closeBtn.style.opacity = '1';
    closeBtn.onmouseout = () => closeBtn.style.opacity = '0.7';
    closeBtn.onclick = () => this.remove(toastId);

    toast.appendChild(content);
    toast.appendChild(closeBtn);
    
    // Add click to dismiss or open feedback widget
    toast.onclick = (e) => {
      if (e.target !== closeBtn) {
        if (data.filename) {
          // Open feedback widget if filename is provided
          this.openFeedbackWidget(data.filename);

        }
        this.remove(toastId);
      }
    };

    this.container!.appendChild(toast);

    // Animate in
    requestAnimationFrame(() => {
      toast.style.opacity = '1';
      toast.style.transform = 'translateX(0)';
    });

    // Auto-dismiss after 5 seconds
    setTimeout(() => {
      if (document.getElementById(toastId)) {
        this.remove(toastId);
      }
    }, 30000);
  }

  private static getBackgroundColor(type: string): string {
    switch (type) {
      case 'success': return '#4caf50';
      case 'warning': return '#ff9800';
      case 'error': return '#f44336';
      case 'info':
      default: return '#2196f3';
    }
  }

  private static getBorderColor(type: string): string {
    switch (type) {
      case 'success': return '#2e7d32';
      case 'warning': return '#f57c00';
      case 'error': return '#d32f2f';
      case 'info':
      default: return '#1976d2';
    }
  }

  private static remove(toastId: string): void {
    const toast = document.getElementById(toastId);
    if (toast) {
      toast.style.opacity = '0';
      toast.style.transform = 'translateX(100%)';
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
      }, 300);
    }
  }

  private static openFeedbackWidget(filename: string): void {
    // Check if feedback widget already exists for this filename
    let floatingFeedback = this.feedbackWidgets.get(filename);
    
    if (!floatingFeedback) {
      // Create new feedback widget if it doesn't exist
      floatingFeedback = new FloatingFeedbackWidget(filename);
      this.feedbackWidgets.set(filename, floatingFeedback);
      
      // Add cleanup when widget is closed
      const originalClose = floatingFeedback.close.bind(floatingFeedback);
      floatingFeedback.close = () => {
        originalClose();
        this.feedbackWidgets.delete(filename);
      };
    }
    
    floatingFeedback.show();
  }
}

export class SSENotificationService {
  private eventSource: EventSource | null = null;
  private isConnected = false;
  private reconnectTimeout: number | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private baseReconnectDelay = 1000;

  constructor() {
    // Initialize toast system
    ToastNotification.init();
  }

  async connect(): Promise<void> {
    if (this.isConnected || this.eventSource) {
      return;
    }

    try {
      const config = await fetchConfig();
      const sseUrl = `${config.server}/events?user_id=${config.id}`;
      this.eventSource = new EventSource(sseUrl);
      
      this.eventSource.onopen = () => {
        console.log('SSE connection established to /events');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        
      };

      this.eventSource.onmessage = (event) => {
        try {
          // Try to parse as JSON first
          const data = JSON.parse(event.data);

          // Handle different message formats
          let notificationData: NotificationData;

          // console.log(data)
          
          if (data && data.student_id === Number(config.id)) {
            // Derive filename from problem_id if available for Feedbackpanel toggle
            const filename = data.problem_id ? `Exercises/ex${String(data.problem_id).padStart(3, '0')}.ipynb` : undefined;
            
            notificationData = {
              message: `You have new ${data.event_type}. Click to View`,
              type: 'info',
              title: 'New Feedback',
              timestamp: data.timestamp,
              filename: filename
            };

            ToastNotification.show(notificationData);

            // If the feedbackpanel is open, hit refresh
            if (filename) {
              const existingWidget = ToastNotification.feedbackWidgets.get(filename);
              if (existingWidget) {
                existingWidget.refreshFeedback();
              }
            }
          }
          
        } catch (error) {
          console.error('Failed to parse SSE message:', error);
        }
      };

      this.eventSource.onerror = (error) => {
        console.error('SSE connection error:', error);
        this.isConnected = false;
        this.handleConnectionError();
      };

      // Listen for specific notification events
      this.eventSource.addEventListener('notify', (event) => {
        try {
          const data: NotificationData = JSON.parse(event.data);
          ToastNotification.show(data);
        } catch (error) {
          console.error('Failed to parse notify event:', error);
        }
      });

    } catch (error) {
      console.error('Failed to establish SSE connection:', error);
      this.handleConnectionError();
    }
  }

  private handleConnectionError(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    
    this.isConnected = false;

    // Show error notification
    ToastNotification.show({
      message: 'Lost connection to /events endpoint',
      type: 'warning',
      title: 'SSE Connection Lost'
    });

    // Attempt to reconnect
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts);
      this.reconnectAttempts++;
      
      console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`);
      
      this.reconnectTimeout = window.setTimeout(async () => {
        await this.connect();
      }, delay);
    } else {
      ToastNotification.show({
        message: 'Failed to reconnect to /events after multiple attempts',
        type: 'error',
        title: 'SSE Connection Failed'
      });
    }
  }

  disconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    
    this.isConnected = false;
    this.reconnectAttempts = 0;
  }

  isConnectionActive(): boolean {
    return this.isConnected;
  }
}

// Global instance
let sseService: SSENotificationService | null = null;

export function getSSEService(): SSENotificationService {
  if (!sseService) {
    sseService = new SSENotificationService();
  }
  return sseService;
}

export async function initializeNotifications(): Promise<void> {
  const service = getSSEService();
  await service.connect();
}

export function cleanupNotifications(): void {
  if (sseService) {
    sseService.disconnect();
    sseService = null;
  }
}