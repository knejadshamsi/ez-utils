<script lang="ts">
  import { Modal, Button, Checkbox } from 'flowbite-svelte';
  import { PlusOutline, CheckOutline, CloseOutline } from 'flowbite-svelte-icons';
  import { populationState, type PersonAttribute } from '$lib/stores/population.svelte';
  import { trackPersonChange } from '$lib/utils/populationChangeTracking';
  import { editingSession } from '$lib/stores/app.svelte.ts';
  import CompactSelect from '../CompactSelect.svelte';
  
  // Get selected person
  const selectedPerson = $derived(
    populationState.selectedPersonId 
      ? populationState.persons.get(populationState.selectedPersonId)
      : null
  );
  
  // Local state for editing attributes
  let editingAttributes = $state<PersonAttribute[]>([]);
  let tempCustomAttribute = $state<{ name: string; type: PersonAttribute['type']; value: string | number | boolean } | null>(null);
  let customAttributeError = $state<string>('');
  
  // Initialize editing attributes when modal opens
  $effect(() => {
    if (populationState.showAttributesModal && selectedPerson) {
      // Deep clone the attributes for editing
      editingAttributes = selectedPerson.attributes.map((attr: PersonAttribute) => ({
        ...attr
      }));
    }
  });
  
  // Attribute type options
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
    
    // Parse value based on type
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
    if (!selectedPerson) return;
    
    // Create updated person with current attributes
    const updatedPerson = {
      ...selectedPerson,
      attributes: editingAttributes
    };
    
    // Update in store
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    console.log('[AttributesModal] Saving attribute changes');
    console.log('[AttributesModal] Updated person:', updatedPerson);
    console.log('[AttributesModal] editingSession.tableName:', editingSession.tableName);
    
    // Track change
    trackPersonChange(updatedPerson, 'update');
  }
  
  function handleCancel() {
    populationState.showAttributesModal = false;
  }
  
  // Common attribute presets with default values
  const commonAttributes = [
    { name: 'age', type: 'java.lang.Integer' as PersonAttribute['type'], defaultValue: 30 },
    { name: 'employed', type: 'java.lang.Boolean' as PersonAttribute['type'], defaultValue: true },
    { name: 'carAvail', type: 'java.lang.String' as PersonAttribute['type'], defaultValue: 'always' },
    { name: 'income', type: 'java.lang.Double' as PersonAttribute['type'], defaultValue: 50000.0 },
    { name: 'hasLicense', type: 'java.lang.Boolean' as PersonAttribute['type'], defaultValue: true },
    { name: 'gender', type: 'java.lang.String' as PersonAttribute['type'], defaultValue: 'male' },
    { name: 'householdSize', type: 'java.lang.Integer' as PersonAttribute['type'], defaultValue: 2 }
  ];
  
  // Check if attribute name already exists
  function isAttributeNameUsed(name: string): boolean {
    return editingAttributes.some(attr => attr.name.toLowerCase() === name.toLowerCase());
  }
  
  // Check if common attribute is already added
  function isCommonAttributeUsed(name: string): boolean {
    return isAttributeNameUsed(name);
  }
  
  // Check if attribute is from common presets
  function isCommonAttribute(name: string): boolean {
    return commonAttributes.some(attr => attr.name.toLowerCase() === name.toLowerCase());
  }
  
  // Get all custom attributes that are currently added
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
      // Remove if exists
      editingAttributes = editingAttributes.filter((_, i) => i !== existingIndex);
    } else {
      // Add with default value
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
    if (!tempCustomAttribute || !tempCustomAttribute.name.trim()) {
      customAttributeError = 'Please enter an attribute name';
      return;
    }
    
    // Check if value is empty or default
    const hasValue = tempCustomAttribute.type === 'java.lang.Boolean' 
      ? true // Boolean always has a value (true/false)
      : tempCustomAttribute.type === 'java.lang.Integer' || tempCustomAttribute.type === 'java.lang.Double'
      ? tempCustomAttribute.value !== 0 && tempCustomAttribute.value !== 0.0
      : tempCustomAttribute.value !== '';
    
    if (!hasValue) {
      customAttributeError = 'Please enter a value for the attribute';
      return;
    }
    
    const nameLower = tempCustomAttribute.name.trim().toLowerCase();
    
    // Check for duplicates
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
    
    // Add the attribute
    editingAttributes = [...editingAttributes, {
      name: tempCustomAttribute.name.trim(),
      type: tempCustomAttribute.type,
      value: tempCustomAttribute.value,
      included: true
    }];
    
    // Clear temp state
    tempCustomAttribute = null;
    customAttributeError = '';
    
    // Save changes immediately
    saveAttributeChanges();
  }
  
  function cancelCustomAttribute() {
    tempCustomAttribute = null;
    customAttributeError = '';
  }
  
  function updateTempCustomValue(value: string) {
    if (!tempCustomAttribute) return;
    
    let parsedValue: string | number | boolean = value;
    
    // Parse value based on type
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
  open={populationState.showAttributesModal}
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
      
      {#if editingAttributes.length === 0 && !tempCustomAttribute}
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
          disabled={tempCustomAttribute !== null}
          class="border-dashed {tempCustomAttribute !== null ? 'opacity-50 cursor-not-allowed' : ''}"
        >
          <PlusOutline class="w-4 h-4 mr-2" />
          Custom
        </Button>
      </div>
    </div>
  </div>
</Modal>