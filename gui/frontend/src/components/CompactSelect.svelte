<script lang="ts">
  export let value: string;
  export let items: Array<{ value: string; name: string }> = [];
  export let onchange: (event: Event) => void = () => {};
  export let disabled: boolean = false;
</script>

<select
  bind:value
  on:change={onchange}
  {disabled}
  class="block py-1 px-2 text-sm rounded-lg border focus:ring-4 focus:outline-none bg-gray-600 text-white border-gray-500 flex-1 h-6 text-xs focus:ring-blue-500 focus:border-blue-500"
>
  {#each items as item}
    <option value={item.value}>{item.name}</option>
  {/each}
</select>

<style>
  select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3E%3Cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3E%3C/svg%3E");
    background-position: right 0.5rem center;
    background-repeat: no-repeat;
    background-size: 1.5em 1.5em;
    padding-right: 2.5rem;
  }
</style>