export type AttachStatus = 'idle' | 'loading' | 'success' | 'error';

export interface AttachState {
  status: AttachStatus;
  errorMessage: string | null;
}
