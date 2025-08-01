<script lang="ts">
  import type { Snippet } from 'svelte';
  
  interface Props {
    name: string;
    value: string;
    onchange?: (value: string) => void;
    children: Snippet;
  }
  
  let { name, value = $bindable(), onchange, children }: Props = $props();
  
  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    onchange?.(value);
  }
</script>

<div class="inline-flex rounded-lg shadow-sm" role="group">
  <div class="contents" onchange={handleChange}>
    {@render children()}
  </div>
</div>