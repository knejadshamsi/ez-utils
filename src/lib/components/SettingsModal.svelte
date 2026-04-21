<script lang="ts">
  import { t } from 'svelte-i18n';
  import { X } from 'lucide-svelte';

  import { sources } from '$lib/stores/data.svelte';
  import { settings, THEMES, LOCALES } from '$lib/stores/ui.svelte';
  import type { ThemeName, LocaleName } from '$lib/stores/ui.svelte';
  import type { CrsPresetKind } from '$lib/components/types';

  const themeNames = Object.keys(THEMES) as ThemeName[];
  const localeNames = Object.keys(LOCALES) as LocaleName[];

  const CRS_PRESETS: CrsPresetKind[] = [
    'montreal_mtm8',
    'tehran_utm39n',
    'paris_lambert93',
    'berlin_etrs89_utm33n',
    'london_bng',
    'new_york_utm18n',
    'tokyo_jgd2011_zone9',
    'sydney_mga94_zone56',
    'zurich_lv95',
  ];

  const crsLocked = $derived(sources.items.length > 0);

  let selectedOption = $state<CrsPresetKind | 'custom'>(
    settings.config.crs.kind === 'custom' ? 'custom' : (settings.config.crs.kind as CrsPresetKind),
  );

  let customDraft = $state({
    projString: settings.config.crs.kind === 'custom' ? settings.config.crs.projString : '',
    lng: settings.config.crs.kind === 'custom' ? settings.config.crs.center[0].toString() : '',
    lat: settings.config.crs.kind === 'custom' ? settings.config.crs.center[1].toString() : '',
    label: settings.config.crs.kind === 'custom' ? settings.config.crs.label : '',
  });

  function onCrsSelectChange(value: string) {
    if (value === 'custom') {
      selectedOption = 'custom';
      return;
    }
    selectedOption = value as CrsPresetKind;
    void settings.setCrs({ kind: value as CrsPresetKind });
  }

  const customValid = $derived(
    customDraft.projString.trim().length > 0
      && customDraft.label.trim().length > 0
      && Number.isFinite(Number(customDraft.lng))
      && Number.isFinite(Number(customDraft.lat)),
  );

  function applyCustom() {
    if (!customValid) return;
    void settings.setCrs({
      kind: 'custom',
      projString: customDraft.projString.trim(),
      center: [Number(customDraft.lng), Number(customDraft.lat)],
      label: customDraft.label.trim(),
    });
  }

  function onBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) settings.hide();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') settings.hide();
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if settings.open}
  <div
    class="fixed inset-0 z-[2700] flex items-center justify-center bg-black/50"
    role="button"
    tabindex="-1"
    aria-label="Close settings"
    onclick={onBackdropClick}
    onkeydown={onKeydown}
  >
    <div class="bg-base-100 text-base-content rounded-lg shadow-xl w-[360px] max-h-[80vh] flex flex-col">
      <div class="flex items-center justify-between px-4 py-3">
        <h2 class="text-sm font-semibold">{$t('settings.title')}</h2>
        <button class="btn btn-ghost btn-xs btn-square" onclick={() => settings.hide()}>
          <X size={14} />
        </button>
      </div>

      <div class="p-4 overflow-y-auto flex-1 space-y-6">
        <div>
          <span class="text-xs font-medium text-base-content/70 mb-2 block">{$t('theme.label')}</span>
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

        <div>
          <span class="text-xs font-medium text-base-content/70 mb-2 block">{$t('locale.label')}</span>
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

        <div>
          <label for="crs-select" class="text-xs font-medium text-base-content/70 mb-2 block">{$t('crs.label')}</label>
          <select
            id="crs-select"
            class="select select-sm select-bordered w-full"
            value={selectedOption}
            disabled={crsLocked}
            onchange={(e) => onCrsSelectChange((e.currentTarget as HTMLSelectElement).value)}
          >
            {#each CRS_PRESETS as preset}
              <option value={preset}>{$t(`crs.${preset}`)}</option>
            {/each}
            <option value="custom">{$t('crs.custom')}</option>
          </select>
          {#if crsLocked}
            <div class="text-xs text-base-content/60 mt-1">{$t('crs.locked_when_sources')}</div>
          {/if}

          {#if selectedOption === 'custom'}
            <div class="mt-3 space-y-2">
              <div>
                <label for="crs-custom-label" class="text-xs text-base-content/70 mb-1 block">{$t('crs.label_input_label')}</label>
                <input
                  id="crs-custom-label"
                  type="text"
                  class="input input-sm input-bordered w-full"
                  value={customDraft.label}
                  disabled={crsLocked}
                  oninput={(e) => (customDraft.label = (e.currentTarget as HTMLInputElement).value)}
                />
              </div>
              <div>
                <label for="crs-custom-proj" class="text-xs text-base-content/70 mb-1 block">{$t('crs.proj_string_label')}</label>
                <textarea
                  id="crs-custom-proj"
                  class="textarea textarea-sm textarea-bordered w-full font-mono text-xs"
                  rows="3"
                  value={customDraft.projString}
                  disabled={crsLocked}
                  oninput={(e) => (customDraft.projString = (e.currentTarget as HTMLTextAreaElement).value)}
                ></textarea>
              </div>
              <div class="flex gap-2">
                <div class="flex-1">
                  <label for="crs-custom-lng" class="text-xs text-base-content/70 mb-1 block">{$t('crs.center_lng_label')}</label>
                  <input
                    id="crs-custom-lng"
                    type="number"
                    step="any"
                    class="input input-sm input-bordered w-full"
                    value={customDraft.lng}
                    disabled={crsLocked}
                    oninput={(e) => (customDraft.lng = (e.currentTarget as HTMLInputElement).value)}
                  />
                </div>
                <div class="flex-1">
                  <label for="crs-custom-lat" class="text-xs text-base-content/70 mb-1 block">{$t('crs.center_lat_label')}</label>
                  <input
                    id="crs-custom-lat"
                    type="number"
                    step="any"
                    class="input input-sm input-bordered w-full"
                    value={customDraft.lat}
                    disabled={crsLocked}
                    oninput={(e) => (customDraft.lat = (e.currentTarget as HTMLInputElement).value)}
                  />
                </div>
              </div>
              <button
                class="btn btn-sm btn-primary w-full"
                disabled={crsLocked || !customValid}
                onclick={applyCustom}
              >{$t('crs.apply_custom')}</button>
            </div>
          {/if}
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label for="autosave-toggle" class="text-xs font-medium text-base-content/70">Autosave</label>
            <input
              id="autosave-toggle"
              type="checkbox"
              class="toggle toggle-sm"
              checked={settings.config.autosaveSettings.enabled}
              onchange={(e) => settings.setAutosaveEnabled((e.currentTarget as HTMLInputElement).checked)}
            />
          </div>
          <div>
            <label for="autosave-interval" class="text-xs font-medium text-base-content/70 mb-2 block">Autosave Interval (minutes)</label>
            <input
              id="autosave-interval"
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
