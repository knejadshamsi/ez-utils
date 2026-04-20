import { mount } from 'svelte'
import './app.css'
import '$lib/i18n'
import { waitLocale } from 'svelte-i18n'
import App from './App.svelte'

waitLocale().then(() => {
  mount(App, {
    target: document.getElementById('app')!,
  })
})
