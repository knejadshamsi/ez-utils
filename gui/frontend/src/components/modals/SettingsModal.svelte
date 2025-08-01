<script lang="ts">
  import { Modal, Button } from "flowbite-svelte";
  import { CogOutline } from "flowbite-svelte-icons";
  import { slide } from "svelte/transition";
  
  interface Props {
    open: boolean;
    onClose: () => void;
  }
  
  let { open, onClose }: Props = $props();
</script>

<Modal bind:open={open} size="lg" autoclose transition={slide}>
  <div class="space-y-6">
    <div class="flex items-center space-x-3">
      <CogOutline class="w-6 h-6 text-gray-500 dark:text-gray-400" />
      <h3 class="text-xl font-semibold text-gray-900 dark:text-white">Settings</h3>
    </div>
    <div class="text-gray-600 dark:text-gray-400">
      <p>Settings will be implemented here.</p>
    </div>
  </div>
  {#snippet footer()}
    <Button color="primary" onclick={onClose}>Close</Button>
  {/snippet}
</Modal>