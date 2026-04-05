<script lang="ts">
  import { t } from 'svelte-i18n';
  import { X } from 'lucide-svelte';
  import { settings, THEMES, LOCALES } from '$lib/stores/ui.svelte';
  import type { ThemeName, LocaleName } from '$lib/stores/ui.svelte';

  const themeNames = Object.keys(THEMES) as ThemeName[];
  const localeNames = Object.keys(LOCALES) as LocaleName[];

  function onBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) settings.hide();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') settings.hide();
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if settings.open}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/50" onclick={onBackdropClick}>
    <div class="bg-base-100 text-base-content rounded-lg shadow-xl w-[340px] max-h-[80vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3">
        <h2 class="text-sm font-semibold">{$t('settings.title')}</h2>
        <button class="btn btn-ghost btn-xs btn-square" onclick={() => settings.hide()}>
          <X size={14} />
        </button>
      </div>

      <!-- Body -->
      <div class="p-4 overflow-y-auto flex-1 space-y-6">
        <!-- Theme -->
        <div>
          <label class="text-xs font-medium text-base-content/70 mb-2 block">{$t('theme.label')}</label>
          <div class="flex gap-2">
            {#each themeNames as name}
              <button
                class="btn btn-sm flex-1"
                class:btn-primary={settings.config.theme === name}
                class:btn-ghost={settings.config.theme !== name}
                onclick={() => settings.setTheme(name)}
              >
                {$t(`theme.${name}`)}
              </button>
            {/each}
          </div>
        </div>

        <!-- Language -->
        <div>
          <label class="text-xs font-medium text-base-content/70 mb-2 block">{$t('locale.label')}</label>
          <div class="flex gap-2">
            {#each localeNames as name}
              <button
                class="btn btn-sm flex-1"
                class:btn-primary={settings.config.locale === name}
                class:btn-ghost={settings.config.locale !== name}
                onclick={() => settings.setLocale(name)}
              >
                {LOCALES[name]}
              </button>
            {/each}
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="text-xs font-medium text-base-content/70">Autosave</label>
            <input
              type="checkbox"
              class="toggle toggle-sm"
              checked={settings.config.autosaveSettings.enabled}
              onchange={(e) => settings.setAutosaveEnabled((e.currentTarget as HTMLInputElement).checked)}
            />
          </div>
          <div>
            <label class="text-xs font-medium text-base-content/70 mb-2 block">Autosave Interval (minutes)</label>
            <input
              type="number"
              min="5"
              max="1440"
              class="input input-sm input-bordered w-full"
              value={settings.config.autosaveSettings.intervalMinutes}
              onchange={(e) => settings.setAutosaveInterval(Number((e.currentTarget as HTMLInputElement).value))}
            />
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
