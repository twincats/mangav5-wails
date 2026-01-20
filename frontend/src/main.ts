import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// General Font
import 'vfonts/Lato.css'
// Monospace Font
import 'vfonts/FiraCode.css'
// Unocss
import 'virtual:uno.css'

createApp(App).use(router).mount('#app')
