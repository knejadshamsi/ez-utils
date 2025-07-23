import { mount } from 'svelte'
import App from './App.svelte'
import './style.css'

function mountApp() {
  const target = document.getElementById('app');
  
  if (!target) {
    return;
  }
  
  try {
    const app = mount(App, {
      target: target
    });
    return app;
  } catch (error) {
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', mountApp);
} else {
  mountApp();
}

export default mountApp;