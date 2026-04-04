<script lang="ts">
  import { t } from 'svelte-i18n';
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
    <div class="bg-base-100 text-base-content rounded-lg shadow-xl w-[620px] min-h-[400px] max-h-[80vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-base-300">
        <h2 class="text-sm font-semibold">{$t('settings.title')}</h2>
        <button class="btn btn-ghost btn-xs btn-square" onclick={() => settings.hide()}>
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
        </button>
      </div>

      <!-- Body: sidebar tabs + content -->
      <div class="flex flex-1 overflow-hidden">
        <!-- Vertical tabs -->
        <div class="flex flex-col border-r border-base-300 w-[120px] shrink-0 py-2">
          <button
            class="px-4 py-2 text-xs font-medium text-left transition-colors {settings.config.activeTab === 'general' ? 'text-primary bg-base-200 border-r-2 border-primary' : 'opacity-50 hover:opacity-80'}"
            onclick={() => settings.setTab('general')}
          >
            {$t('settings.general')}
          </button>
          <button
            class="px-4 py-2 text-xs font-medium text-left transition-colors {settings.config.activeTab === 'display' ? 'text-primary bg-base-200 border-r-2 border-primary' : 'opacity-50 hover:opacity-80'}"
            onclick={() => settings.setTab('display')}
          >
            {$t('settings.display')}
          </button>
          <button
            class="px-4 py-2 text-xs font-medium text-left transition-colors {settings.config.activeTab === 'sources' ? 'text-primary bg-base-200 border-r-2 border-primary' : 'opacity-50 hover:opacity-80'}"
            onclick={() => settings.setTab('sources')}
          >
            {$t('settings.sources')}
          </button>
        </div>

        <!-- Content -->
        <div class="p-4 overflow-y-auto flex-1">
          {#if settings.config.activeTab === 'display'}
            <div class="space-y-6">
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
            </div>
          {:else if settings.config.activeTab === 'general'}
            <p class="text-xs text-base-content/50">{$t('settings.general_placeholder')}</p>
          {:else if settings.config.activeTab === 'sources'}
            <p class="text-xs text-base-content/50">{$t('settings.sources_placeholder')}</p>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
