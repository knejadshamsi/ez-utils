import { mount } from 'svelte'
import App from './App.svelte'
import './style.css'

// Wait for DOM to be ready
function mountApp() {
  console.log('Attempting to mount Svelte app...');
  
  const target = document.getElementById('app');
  console.log('Target element:', target);
  
  if (!target) {
    console.error('Could not find app element!');
    return;
  }
  
  try {
    // Svelte 5 mount API
    const app = mount(App, {
      target: target
    });
    console.log('Svelte app mounted successfully!', app);
    return app;
  } catch (error) {
    console.error('Error mounting Svelte app:', error);
  }
}

// Ensure DOM is loaded before mounting
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', mountApp);
} else {
  mountApp();
}

export default mountApp;