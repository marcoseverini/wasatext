import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import axios from './services/axios.js'; // Importiamo la nostra istanza
import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)

// Configurazione globale stile Fantastic Coffee
app.config.globalProperties.$axios = axios;

app.component("ErrorMsg", ErrorMsg);
app.component("LoadingSpinner", LoadingSpinner);

app.use(router)
app.mount('#app')