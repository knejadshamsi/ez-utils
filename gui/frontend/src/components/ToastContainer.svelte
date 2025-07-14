<script lang="ts">
  import { Toast } from 'flowbite-svelte';
  import { CheckCircleOutline, CloseCircleOutline, ExclamationCircleOutline, InfoCircleOutline } from 'flowbite-svelte-icons';
  import { toastStore, removeToast, getToastColor } from '../lib/toast.svelte';
  import { fly } from 'svelte/transition';
  
  function getIcon(type: string) {
    switch (type) {
      case 'success':
        return CheckCircleOutline;
      case 'error':
        return CloseCircleOutline;
      case 'warning':
        return ExclamationCircleOutline;
      case 'info':
        return InfoCircleOutline;
      default:
        return InfoCircleOutline;
    }
  }
</script>

<div class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 space-y-2">
  {#each toastStore.toasts as toast (toast.id)}
    <div transition:fly={{ y: -50, duration: 300 }}>
      <Toast 
        color={getToastColor(toast.type)}
        class="mb-2"
        onclose={() => removeToast(toast.id)}
      >
        {#snippet icon()}
          <svelte:component this={getIcon(toast.type)} class="w-5 h-5" />
        {/snippet}
        <span class="text-sm">{toast.message}</span>
      </Toast>
    </div>
  {/each}
</div>