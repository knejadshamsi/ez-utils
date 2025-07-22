<script lang="ts">
  import { Modal, Button } from 'flowbite-svelte';
  import { ExclamationCircleOutline } from 'flowbite-svelte-icons';
  import { slide } from 'svelte/transition';
  
  interface Props {
    open: boolean;
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    onConfirm: () => void;
    onCancel: () => void;
  }
  
  let { 
    open = $bindable(),
    title,
    message, 
    confirmText = "Continue",
    cancelText = "Cancel",
    onConfirm,
    onCancel 
  }: Props = $props();
</script>

<Modal bind:open size="sm" autoclose={false} transition={slide}>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-orange-400" />
    <h3 class="mb-2 text-lg font-semibold text-gray-700 dark:text-gray-300">{title}</h3>
    <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{message}</p>
    
    <div class="flex justify-center gap-3">
      <Button color="orange" onclick={() => { onConfirm(); open = false; }}>
        {confirmText}
      </Button>
      <Button color="light" onclick={() => { onCancel(); open = false; }}>
        {cancelText}
      </Button>
    </div>
  </div>
</Modal>