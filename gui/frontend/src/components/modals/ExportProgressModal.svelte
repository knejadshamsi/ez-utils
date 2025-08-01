<script lang="ts">
  import { Modal, Spinner } from "flowbite-svelte";
  import { exportProgress } from "$lib/stores/exportProgress.svelte";
</script>

<Modal 
  open={exportProgress.isExporting}
  permanent
  dismissable={false}
  outsideclose={false}
  class="!p-0"
  size="sm"
  placement="center"
>
  <div class="flex items-center justify-center gap-3 p-6">
    <Spinner size="6" color="blue" />
    <span class="text-gray-700 dark:text-gray-300">
      Exporting {exportProgress.current.toLocaleString()}/{exportProgress.total > 0 ? exportProgress.total.toLocaleString() : '?'} elements...
    </span>
  </div>
</Modal>