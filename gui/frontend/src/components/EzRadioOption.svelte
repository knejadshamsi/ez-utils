<script lang="ts">
  import type { Snippet } from 'svelte';
  
  interface Props {
    value: string;
    name: string;
    checked?: boolean;
    children: Snippet;
  }
  
  let { value, name, checked = false, children }: Props = $props();
</script>

<label class="ez-radio-option relative inline-flex items-center justify-center py-2.5 px-5 text-sm font-medium text-gray-600 bg-white border border-gray-300 cursor-pointer transition-all duration-150 ease-in-out hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-800 dark:border-gray-600 dark:hover:bg-gray-700">
  <input
    type="radio"
    {name}
    {value}
    {checked}
    class="absolute w-px h-px p-0 -m-px overflow-hidden whitespace-nowrap border-0"
    style="clip: rect(0, 0, 0, 0);"
  />
  {@render children()}
</label>

<style>
  /* Border radius for first and last child */
  .ez-radio-option:first-child {
    border-radius: 0.5rem 0 0 0.5rem;
  }
  
  .ez-radio-option:last-child {
    border-radius: 0 0.5rem 0.5rem 0;
  }
  
  /* Remove right border except for last child */
  .ez-radio-option:not(:last-child) {
    border-right: 0;
  }
  
  /* Checked state styling */
  .ez-radio-option:has(input:checked) {
    color: white;
    background-color: rgb(29 78 216); /* blue-700 - Flowbite primary */
    border-color: rgb(29 78 216);
    z-index: 10;
  }
  
  /* Dark mode checked state */
  :global(.dark) .ez-radio-option:has(input:checked) {
    background-color: rgb(26 86 219); /* blue-600 - Flowbite primary dark */
    border-color: rgb(26 86 219);
  }
</style>