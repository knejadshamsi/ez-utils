export type ToastType = 'success' | 'error' | 'warning' | 'info';
export type ToastColor = 'green' | 'red' | 'orange' | 'blue' | 'gray';

interface ToastItem {
  id: string;
  type: ToastType;
  message: string;
  duration: number;
}

// Create reactive toast store
export const toastStore = $state<{
  toasts: ToastItem[];
}>({
  toasts: []
});

// Toast functions
export function showToast(type: ToastType, message: string, duration: number = 3000) {
  const id = `toast-${Date.now()}-${Math.random()}`;
  const toast: ToastItem = {
    id,
    type,
    message,
    duration
  };
  
  toastStore.toasts = [...toastStore.toasts, toast];
  
  // Auto-remove after duration
  if (duration > 0) {
    setTimeout(() => {
      removeToast(id);
    }, duration);
  }
  
  return id;
}

export function removeToast(id: string) {
  toastStore.toasts = toastStore.toasts.filter(t => t.id !== id);
}

export function showSuccess(message: string, duration?: number) {
  return showToast('success', message, duration);
}

export function showError(message: string, duration?: number) {
  return showToast('error', message, duration);
}

export function showWarning(message: string, duration?: number) {
  return showToast('warning', message, duration);
}

export function showInfo(message: string, duration?: number) {
  return showToast('info', message, duration);
}

// Get color based on type
export function getToastColor(type: ToastType): ToastColor {
  switch (type) {
    case 'success':
      return 'green';
    case 'error':
      return 'red';
    case 'warning':
      return 'orange';
    case 'info':
      return 'blue';
    default:
      return 'gray';
  }
}