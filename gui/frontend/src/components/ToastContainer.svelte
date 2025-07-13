<script lang="ts">
  import { Toast } from 'flowbite-svelte';
  import { CheckCircleOutline, CloseCircleOutline, ExclamationCircleOutline, InfoCircleOutline, CloseOutline } from 'flowbite-svelte-icons';
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

<div class="fixed top-20 right-4 z-50 space-y-2">
  {#each toastStore.toasts as toast (toast.id)}
    <div transition:fly={{ x: 100, duration: 300 }}>
      <Toast 
        color={getToastColor(toast.type)}
        class="mb-2"
      >
        <svelte:component this={getIcon(toast.type)} slot="icon" class="w-5 h-5" />
        <span class="text-sm">{toast.message}</span>
        <button
          type="button"
          class="ml-auto -mx-1.5 -my-1.5 bg-white text-gray-400 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 hover:bg-gray-100 inline-flex h-8 w-8 dark:bg-gray-800 dark:text-gray-500 dark:hover:text-white dark:hover:bg-gray-700"
          aria-label="Close"
          onclick={() => removeToast(toast.id)}
        >
          <CloseOutline class="w-5 h-5" />
        </button>
      </Toast>
    </div>
  {/each}
</div>