export interface ConfirmationOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel?: () => void;
}

export const confirmationState = $state<{
  isOpen: boolean;
  options: ConfirmationOptions | null;
}>({
  isOpen: false,
  options: null
});

export function showConfirmation(options: ConfirmationOptions) {
  confirmationState.options = options;
  confirmationState.isOpen = true;
  
  return new Promise<boolean>((resolve) => {
    const originalOnConfirm = options.onConfirm;
    const originalOnCancel = options.onCancel;
    
    confirmationState.options = {
      ...options,
      onConfirm: () => {
        originalOnConfirm();
        confirmationState.isOpen = false;
        confirmationState.options = null;
        resolve(true);
      },
      onCancel: () => {
        originalOnCancel?.();
        confirmationState.isOpen = false;
        confirmationState.options = null;
        resolve(false);
      }
    };
  });
}

export function hideConfirmation() {
  confirmationState.isOpen = false;
  confirmationState.options = null;
}