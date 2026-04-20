<script lang="ts">
  import AttributeTableEditor from '$lib/components/attributes/AttributeTableEditor.svelte';
  import { population } from '$lib/stores/population.svelte';
</script>

<AttributeTableEditor
  rows={population.attributeRows}
  onUpdateName={(index, value) => population.updateAttributeName(index, value)}
  onUpdateType={(index, value) => population.updateAttributeType(index, value)}
  onUpdateValue={(index, value) => population.updateAttributeValue(index, value)}
  onRemove={(index) => population.removeAttributeRow(index)}
/>
