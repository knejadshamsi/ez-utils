<script lang="ts">
  import { population } from '$lib/stores/population.svelte';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
</script>

<ConfirmModal
  open={population.deletePromptOpen}
  title="Delete Person"
  onConfirm={() => void population.confirmDeletePerson()}
  onCancel={() => population.cancelDeletePerson()}
>
  <p>Delete person {population.deleteTarget?.personId}? This cannot be undone.</p>
</ConfirmModal>
