<script lang="ts">
  import { Modal, Button, Checkbox } from 'flowbite-svelte';
  import { PlusOutline, CheckOutline, CloseOutline } from 'flowbite-svelte-icons';
  import { populationState } from '@workflow/population/state.svelte';
  import type { PersonAttribute } from '@workflow/population/types';
  import { trackPersonChange } from '@workflow/population/populationChangeTracking';
  import CompactSelect from '../CompactSelect.svelte';
  
  const selectedPerson = $derived(
    populationState.sidebarInteraction === 'EDITING_ATTRIBUTES' 
      ? populationState.persons.get(populationState.selectedPersonId!)! 
      : null
  );
  
  let editingAttributes = $state<PersonAttribute[]>([]);
  let tempCustomAttribute = $state<{ name: string; type: PersonAttribute['type']; value: string | number | boolean } | null>(null);
  let customAttributeError = $state<string>('');
  
  $effect(() => {
    if (populationState.sidebarInteraction === 'EDITING_ATTRIBUTES') {
      editingAttributes = selectedPerson!.attributes.map((attr: PersonAttribute) => ({
        ...attr
      }));
    }
  });
  
  const attributeTypes = [
    { value: 'java.lang.String', name: 'String' },
    { value: 'java.lang.Integer', name: 'Integer' },
    { value: 'java.lang.Double', name: 'Double' },
    { value: 'java.lang.Boolean', name: 'Boolean' }
  ];
  
  
  function removeAttribute(name: string) {
    const confirmRemove = confirm(`Remove attribute "${name}"?`);
    if (confirmRemove) {
      editingAttributes = editingAttributes.filter(attr => attr.name !== name);
      saveAttributeChanges();
    }
  }
  
  function updateAttributeValue(index: number, value: string) {
    const attr = editingAttributes[index];
    let parsedValue: string | number | boolean = value;
    
    if (attr.type === 'java.lang.Integer') {
      parsedValue = parseInt(value, 10) || 0;
    } else if (attr.type === 'java.lang.Double') {
      parsedValue = parseFloat(value) || 0.0;
    } else if (attr.type === 'java.lang.Boolean') {
      parsedValue = value.toLowerCase() === 'true';
    }
    
    editingAttributes[index].value = parsedValue;
    saveAttributeChanges();
  }
  
  function saveAttributeChanges() {
    const updatedPerson = {
      ...selectedPerson!,
      attributes: editingAttributes
    };
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    
    trackPersonChange(updatedPerson, 'update');
  }
  
  function handleCancel() {
    populationState.sidebarInteraction = 'NORMAL';
  }
  
  const commonAttributes = [
    { name: 'age', type: 'java.lang.Integer' as PersonAttribute['type'], defaultValue: 30 },
    { name: 'employed', type: 'java.lang.Boolean' as PersonAttribute['type'], defaultValue: true },
    { name: 'carAvail', type: 'java.lang.String' as PersonAttribute['type'], defaultValue: 'always' },
    { name: 'income', type: 'java.lang.Double' as PersonAttribute['type'], defaultValue: 50000.0 },
    { name: 'hasLicense', type: 'java.lang.Boolean' as PersonAttribute['type'], defaultValue: true },
    { name: 'gender', type: 'java.lang.String' as PersonAttribute['type'], defaultValue: 'male' },
    { name: 'householdSize', type: 'java.lang.Integer' as PersonAttribute['type'], defaultValue: 2 }
  ];
  
  function isAttributeNameUsed(name: string): boolean {
    return editingAttributes.some(attr => attr.name.toLowerCase() === name.toLowerCase());
  }
  
  function isCommonAttributeUsed(name: string): boolean {
    return isAttributeNameUsed(name);
  }
  
  function isCommonAttribute(name: string): boolean {
    return commonAttributes.some(attr => attr.name.toLowerCase() === name.toLowerCase());
  }
  
  const addedCustomAttributes = $derived(
    editingAttributes
      .filter(attr => !isCommonAttribute(attr.name))
      .map(attr => ({ name: attr.name, type: attr.type as PersonAttribute['type'] }))
  );
  
  function toggleCommonAttribute(preset: { name: string; type: PersonAttribute['type']; defaultValue: any }) {
    const existingIndex = editingAttributes.findIndex(attr => 
      attr.name.toLowerCase() === preset.name.toLowerCase()
    );
    
    if (existingIndex >= 0) {
      editingAttributes = editingAttributes.filter((_, i) => i !== existingIndex);
    } else {
      editingAttributes = [...editingAttributes, {
        name: preset.name,
        type: preset.type,
        value: preset.defaultValue,
        included: true
      }];
    }
    saveAttributeChanges();
  }
  
  function startAddingCustomAttribute() {
    tempCustomAttribute = {
      name: '',
      type: 'java.lang.String',
      value: ''
    };
    customAttributeError = '';
  }
  
  function confirmCustomAttribute() {
    if (!tempCustomAttribute!.name.trim()) {
      customAttributeError = 'Please enter an attribute name';
      return;
    }
    
    const hasValue = tempCustomAttribute.type === 'java.lang.Boolean' 
      ? true
      : tempCustomAttribute.type === 'java.lang.Integer' || tempCustomAttribute.type === 'java.lang.Double'
      ? tempCustomAttribute.value !== 0 && tempCustomAttribute.value !== 0.0
      : tempCustomAttribute.value !== '';
    
    if (!hasValue) {
      customAttributeError = 'Please enter a value for the attribute';
      return;
    }
    
    const nameLower = tempCustomAttribute.name.trim().toLowerCase();
    
    const isDuplicate = editingAttributes.some(attr => 
      attr.name.toLowerCase() === nameLower
    );
    
    const isCommonName = commonAttributes.some(attr => 
      attr.name.toLowerCase() === nameLower
    );
    
    if (isDuplicate || isCommonName) {
      customAttributeError = `"${tempCustomAttribute.name}" already exists`;
      return;
    }
    
    editingAttributes = [...editingAttributes, {
      name: tempCustomAttribute.name.trim(),
      type: tempCustomAttribute.type,
      value: tempCustomAttribute.value,
      included: true
    }];
    
    tempCustomAttribute = null;
    customAttributeError = '';
    
    saveAttributeChanges();
  }
  
  function cancelCustomAttribute() {
    tempCustomAttribute = null;
    customAttributeError = '';
  }
  
  function updateTempCustomValue(value: string) {
    
    let parsedValue: string | number | boolean = value;
    
    if (tempCustomAttribute.type === 'java.lang.Integer') {
      parsedValue = parseInt(value, 10) || 0;
    } else if (tempCustomAttribute.type === 'java.lang.Double') {
      parsedValue = parseFloat(value) || 0.0;
    } else if (tempCustomAttribute.type === 'java.lang.Boolean') {
      parsedValue = value.toLowerCase() === 'true';
    }
    
    tempCustomAttribute.value = parsedValue;
  }
  
  
</script>

<Modal
  title="Person Attributes - {selectedPerson?.id || ''}"
  open={populationState.sidebarInteraction === 'EDITING_ATTRIBUTES'}
  onclose={handleCancel}
  class="min-w-[600px]"
>
  <div class="space-y-4">
    <!-- Attribute list -->
    <div class="space-y-2 max-h-96 overflow-y-auto">
      {#each editingAttributes as attr, index}
        <div class="flex items-center gap-2">
          <Checkbox bind:checked={attr.included} onchange={saveAttributeChanges} />
          
          {#if isCommonAttribute(attr.name)}
            <div class="block py-1.5 px-3 text-sm rounded-lg border bg-gray-700 text-gray-300 border-gray-600 flex-1 h-8">
              {attr.name}
            </div>
          {:else}
            <div class="block py-1.5 px-3 text-sm rounded-lg border bg-gray-600 text-white border-gray-500 flex-1 h-8">
              {attr.name}
            </div>
          {/if}
          
          {#if isCommonAttribute(attr.name)}
            <div class="block py-1.5 px-3 text-sm rounded-lg border bg-gray-700 text-gray-300 border-gray-600 w-32 h-8">
              {attributeTypes.find(t => t.value === attr.type)?.name || attr.type}
            </div>
          {:else}
            <CompactSelect
              bind:value={attr.type}
              items={attributeTypes}
              size="md"
              onchange={saveAttributeChanges}
            />
          {/if}
          
          {#if attr.type === 'java.lang.Boolean'}
            <CompactSelect
              value={String(attr.value)}
              onchange={(e) => updateAttributeValue(index, (e.target as HTMLSelectElement).value)}
              items={[
                { value: 'true', name: 'true' },
                { value: 'false', name: 'false' }
              ]}
              size="md"
            />
          {:else}
            <input
              placeholder="Value"
              value={String(attr.value)}
              oninput={(e) => updateAttributeValue(index, (e.target as HTMLInputElement).value)}
              type={attr.type === 'java.lang.Integer' || attr.type === 'java.lang.Double' ? 'number' : 'text'}
              class="block py-1.5 px-3 text-sm rounded-lg border focus:ring-4 focus:outline-none bg-gray-600 text-white border-gray-500 w-32 h-8 focus:ring-blue-500 focus:border-blue-500"
            />
          {/if}
          
        </div>
      {/each}
      
      {#if editingAttributes.length === 0}
        <p class="text-gray-500 text-center py-4">No attributes defined</p>
      {/if}
      
      <!-- Temporary custom attribute form -->
      {#if tempCustomAttribute}
        <div class="flex items-center gap-2 p-3 bg-blue-900/20 border border-blue-600 rounded-lg">
          <input
            placeholder="Attribute name"
            bind:value={tempCustomAttribute.name}
            oninput={() => customAttributeError = ''}
            class="block py-1.5 px-3 text-sm rounded-lg border focus:ring-4 focus:outline-none bg-gray-600 text-white border-gray-500 flex-1 h-8 focus:ring-blue-500 focus:border-blue-500"
          />
          
          <CompactSelect
            bind:value={tempCustomAttribute.type}
            items={attributeTypes}
            size="md"
          />
          
          {#if tempCustomAttribute.type === 'java.lang.Boolean'}
            <CompactSelect
              value={String(tempCustomAttribute.value)}
              onchange={(e) => updateTempCustomValue((e.target as HTMLSelectElement).value)}
              items={[
                { value: 'true', name: 'true' },
                { value: 'false', name: 'false' }
              ]}
              size="md"
            />
          {:else}
            <input
              placeholder="Value"
              value={String(tempCustomAttribute.value)}
              oninput={(e) => updateTempCustomValue((e.target as HTMLInputElement).value)}
              type={tempCustomAttribute.type === 'java.lang.Integer' || tempCustomAttribute.type === 'java.lang.Double' ? 'number' : 'text'}
              class="block py-1.5 px-3 text-sm rounded-lg border focus:ring-4 focus:outline-none bg-gray-600 text-white border-gray-500 w-32 h-8 focus:ring-blue-500 focus:border-blue-500"
            />
          {/if}
          
          <Button
            size="xs"
            color="green"
            class="!p-1.5"
            onclick={confirmCustomAttribute}
            title="Confirm"
          >
            <CheckOutline class="w-4 h-4" />
          </Button>
          
          <Button
            size="xs"
            color="red"
            class="!p-1.5"
            onclick={cancelCustomAttribute}
            title="Cancel"
          >
            <CloseOutline class="w-4 h-4" />
          </Button>
        </div>
        
        {#if customAttributeError}
          <p class="text-red-400 text-sm mt-1 ml-3">{customAttributeError}</p>
        {/if}
      {/if}
    </div>
    
    <!-- Common attributes, custom attributes, and add custom button -->
    <div class="border-t pt-4">
      <div class="flex flex-wrap gap-2">
        {#each commonAttributes as preset}
          <Button
            size="sm"
            color={isCommonAttributeUsed(preset.name) ? 'green' : 'alternative'}
            onclick={() => toggleCommonAttribute(preset)}
            class="relative"
          >
            {#if isCommonAttributeUsed(preset.name)}
              <CheckOutline class="w-3 h-3 mr-1" />
            {/if}
            {preset.name}
          </Button>
        {/each}
        
        <!-- Custom attributes -->
        {#each addedCustomAttributes as customAttr}
          <Button
            size="sm"
            color="purple"
            onclick={() => removeAttribute(customAttr.name)}
            class="relative"
          >
            <CheckOutline class="w-3 h-3 mr-1" />
            {customAttr.name}
          </Button>
        {/each}
        
        <Button
          size="sm"
          color="primary"
          onclick={startAddingCustomAttribute}
          disabled={!!tempCustomAttribute}
          class="border-dashed {tempCustomAttribute !== null ? 'opacity-50 cursor-not-allowed' : ''}"
        >
          <PlusOutline class="w-4 h-4 mr-2" />
          Custom
        </Button>
      </div>
    </div>
  </div>
</Modal>