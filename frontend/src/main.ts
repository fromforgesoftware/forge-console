import './assets/index.css';

import { createApp } from 'vue';
import { createPinia } from 'pinia';

import App from './App.vue';
import router from './app/core/router';
import { registerSharedSingletons } from './app/features/console/application/system';

// Grafana-style plugin loading runs on SystemJS: register the host's own copies
// of the shared singletons (vue, vue-router, pinia, the forge kits, the console
// plugin contract) into the SystemJS import map BEFORE any runtime plugin
// module.js is imported, so each plugin's externalised imports resolve to the
// host's live instances (one Vue, one pinia root, one router) instead of a
// second bundled copy. Must run before the router/guards can lazily load a
// plugin module.
registerSharedSingletons();

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.mount('#app');
